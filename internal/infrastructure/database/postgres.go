package database

import (
	"fmt"

	"github.com/davidrdsilva/blog-api/internal/domain/models"
	"github.com/davidrdsilva/blog-api/internal/infrastructure/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultCategoryName is the seed category used to backfill posts that pre-date
// the introduction of the category column.
const defaultCategoryName = "News"

const WhitenestCategoryName = "Whitenest"

const DraftsCategoryName = "Drafts"

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(dsn string, log *logging.Logger) (*gorm.DB, error) {
	log.Info("Connecting to PostgreSQL database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Use custom logger instead
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL")

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Info("Database connection pool configured")

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *gorm.DB, log *logging.Logger) error {
	log.Info("Running database migrations...")

	// Categories and tags must exist before we can wire them into posts.
	if err := db.AutoMigrate(&models.Category{}, &models.Tag{}, &models.Character{}); err != nil {
		return fmt.Errorf("failed to migrate categories/tags/characters: %w", err)
	}

	// Use an explicit join model with its own UUID primary key so the join row
	// is addressable on its own (matches the spec's posts_tags ER diagram).
	if err := db.SetupJoinTable(&models.Post{}, "Tags", &models.PostsTag{}); err != nil {
		return fmt.Errorf("failed to setup posts_tags join: %w", err)
	}

	// Same pattern for characters: explicit join model so we can store extra
	// columns on the relation (`position` for cast ordering).
	if err := db.SetupJoinTable(&models.Post{}, "Characters", &models.PostsCharacter{}); err != nil {
		return fmt.Errorf("failed to setup posts_characters join: %w", err)
	}

	// Stage the category column on posts before AutoMigrating Post, because the
	// model declares category_id as NOT NULL — running AutoMigrate against an
	// existing posts table with rows would fail if the column were added with a
	// NOT NULL constraint up front.
	if err := stageCategoryColumnOnPosts(db, log); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Post{}, &models.Comment{}); err != nil {
		return fmt.Errorf("failed to migrate posts/comments: %w", err)
	}

	if err := seedWhitenestCategory(db, log); err != nil {
		return err
	}
	if err := stageWhitenestChapterColumn(db, log); err != nil {
		return err
	}
	if err := stageInternalCategoryColumn(db, log); err != nil {
		return err
	}
	if err := seedDraftsCategory(db, log); err != nil {
		return err
	}

	log.Info("Database migrations completed successfully")

	if err := createIndexes(db, log); err != nil {
		return err
	}

	return nil
}

func seedWhitenestCategory(db *gorm.DB, log *logging.Logger) error {
	var count int64
	if err := db.Model(&models.Category{}).Where("name = ?", WhitenestCategoryName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check Whitenest category: %w", err)
	}
	if count > 0 {
		return nil
	}
	if err := db.Create(&models.Category{Name: WhitenestCategoryName}).Error; err != nil {
		return fmt.Errorf("failed to create Whitenest category: %w", err)
	}
	log.Info("Seeded Whitenest category", logging.F("name", WhitenestCategoryName))
	return nil
}

// stageInternalCategoryColumn adds categories.is_internal as a boolean default
// false on existing databases. AutoMigrate handles fresh databases; this guards
// existing instances that need the column added without breaking inserts.
func stageInternalCategoryColumn(db *gorm.DB, log *logging.Logger) error {
	if err := db.Exec(
		`ALTER TABLE categories ADD COLUMN IF NOT EXISTS is_internal BOOLEAN NOT NULL DEFAULT FALSE`,
	).Error; err != nil {
		return fmt.Errorf("failed to add categories.is_internal: %w", err)
	}
	log.Info("categories.is_internal column staged")
	return nil
}

// seedDraftsCategory inserts the "Drafts" row (is_internal=true) if it doesn't
// already exist. Posts assigned to this category are hidden from public feeds.
func seedDraftsCategory(db *gorm.DB, log *logging.Logger) error {
	var count int64
	if err := db.Model(&models.Category{}).Where("name = ?", DraftsCategoryName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check Drafts category: %w", err)
	}
	if count > 0 {
		// Defensive: ensure the row is flagged internal even if it was created
		// before the column existed.
		if err := db.Model(&models.Category{}).
			Where("name = ? AND is_internal = ?", DraftsCategoryName, false).
			Update("is_internal", true).Error; err != nil {
			return fmt.Errorf("failed to mark Drafts category as internal: %w", err)
		}
		return nil
	}
	if err := db.Create(&models.Category{Name: DraftsCategoryName, IsInternal: true}).Error; err != nil {
		return fmt.Errorf("failed to create Drafts category: %w", err)
	}
	log.Info("Seeded Drafts category", logging.F("name", DraftsCategoryName))
	return nil
}

// stageWhitenestChapterColumn (re)asserts the column and its partial unique
// index. AutoMigrate handles fresh databases; this guards existing instances
// picking up the field for the first time.
func stageWhitenestChapterColumn(db *gorm.DB, log *logging.Logger) error {
	if err := db.Exec(
		`ALTER TABLE posts ADD COLUMN IF NOT EXISTS whitenest_chapter_number INTEGER`,
	).Error; err != nil {
		return fmt.Errorf("failed to add posts.whitenest_chapter_number: %w", err)
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_whitenest_chapter_number
		ON posts(whitenest_chapter_number)
		WHERE whitenest_chapter_number IS NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to create whitenest_chapter_number unique index: %w", err)
	}

	log.Info("posts.whitenest_chapter_number staged migration applied")
	return nil
}

// stageCategoryColumnOnPosts adds posts.category_id in three idempotent steps:
//  1. add the column as NULL (safe on a populated table),
//  2. seed the default "News" category and backfill existing rows,
//  3. flip the column to NOT NULL and add the FK to categories.
//
// Each step is guarded so re-running migrations on an already-migrated database
// is a no-op.
func stageCategoryColumnOnPosts(db *gorm.DB, log *logging.Logger) error {
	// Skip the whole staged dance if there is no posts table yet — the next
	// AutoMigrate will create it with the column already in its final shape.
	hasPosts := db.Migrator().HasTable(&models.Post{})
	if !hasPosts {
		return seedDefaultCategory(db, log)
	}

	// Step 1: add the column as nullable (idempotent).
	if err := db.Exec(`ALTER TABLE posts ADD COLUMN IF NOT EXISTS category_id INTEGER`).Error; err != nil {
		return fmt.Errorf("failed to add posts.category_id: %w", err)
	}

	// Step 2: seed the default category and backfill any rows still missing it.
	if err := seedDefaultCategory(db, log); err != nil {
		return err
	}

	var defaultCat models.Category
	if err := db.Where("name = ?", defaultCategoryName).First(&defaultCat).Error; err != nil {
		return fmt.Errorf("failed to load default category: %w", err)
	}

	if err := db.Exec(
		`UPDATE posts SET category_id = ? WHERE category_id IS NULL`,
		defaultCat.ID,
	).Error; err != nil {
		return fmt.Errorf("failed to backfill posts.category_id: %w", err)
	}

	// Step 3: enforce NOT NULL + FK now that no rows are missing the value.
	if err := db.Exec(`ALTER TABLE posts ALTER COLUMN category_id SET NOT NULL`).Error; err != nil {
		return fmt.Errorf("failed to set NOT NULL on posts.category_id: %w", err)
	}

	if err := db.Exec(`
		ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk_posts_category;
		ALTER TABLE posts ADD CONSTRAINT fk_posts_category
			FOREIGN KEY (category_id) REFERENCES categories(id);
	`).Error; err != nil {
		return fmt.Errorf("failed to set FK on posts.category_id: %w", err)
	}

	log.Info("posts.category_id staged migration applied")
	return nil
}

// seedDefaultCategory inserts the "News" row if it doesn't already exist.
func seedDefaultCategory(db *gorm.DB, log *logging.Logger) error {
	var count int64
	if err := db.Model(&models.Category{}).Where("name = ?", defaultCategoryName).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check default category: %w", err)
	}
	if count > 0 {
		return nil
	}
	if err := db.Create(&models.Category{Name: defaultCategoryName}).Error; err != nil {
		return fmt.Errorf("failed to create default category: %w", err)
	}
	log.Info("Seeded default category", logging.F("name", defaultCategoryName))
	return nil
}

// createIndexes creates database indexes for optimized queries
func createIndexes(db *gorm.DB, log *logging.Logger) error {
	log.Info("Creating database indexes...")

	// Index on date for sorting
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_date ON posts(date DESC)").Error; err != nil {
		return fmt.Errorf("failed to create date index: %w", err)
	}

	// Index on author for filtering
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author)").Error; err != nil {
		return fmt.Errorf("failed to create author index: %w", err)
	}

	// Index on total_views for the most-viewed query.
	if err := db.Exec(
		"CREATE INDEX IF NOT EXISTS idx_posts_total_views ON posts(total_views DESC)",
	).Error; err != nil {
		return fmt.Errorf("failed to create total_views index: %w", err)
	}

	// Index on post_id for comments
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id)").Error; err != nil {
		return fmt.Errorf("failed to create post_id index: %w", err)
	}

	// Full-text search index on searchable fields
	// Create a computed column for full-text search
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_posts_search ON posts
		USING GIN(to_tsvector('english',
			title || ' ' || COALESCE(subtitle, '') || ' ' || description
		))
	`).Error; err != nil {
		return fmt.Errorf("failed to create search index: %w", err)
	}

	// Ensure the comments FK has ON DELETE CASCADE.
	// AutoMigrate creates the constraint on fresh databases, but won't alter an
	// existing one, so we re-create it explicitly to handle running instances.
	if err := db.Exec(`
		ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_posts_comments;
		ALTER TABLE comments ADD CONSTRAINT fk_posts_comments
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to set cascade delete on comments: %w", err)
	}

	// Cascade post deletes through the join table so removing a post doesn't
	// leave dangling posts_tags rows.
	if err := db.Exec(`
		ALTER TABLE posts_tags DROP CONSTRAINT IF EXISTS fk_posts_tags_post;
		ALTER TABLE posts_tags ADD CONSTRAINT fk_posts_tags_post
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;
		ALTER TABLE posts_tags DROP CONSTRAINT IF EXISTS fk_posts_tags_tag;
		ALTER TABLE posts_tags ADD CONSTRAINT fk_posts_tags_tag
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to set cascade on posts_tags: %w", err)
	}

	// Prevent duplicate (post, tag) pairs.
	if err := db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_tags_post_tag ON posts_tags(post_id, tag_id)`,
	).Error; err != nil {
		return fmt.Errorf("failed to create posts_tags unique index: %w", err)
	}

	// Cascade post and character deletes through the cast join table.
	if err := db.Exec(`
		ALTER TABLE posts_characters DROP CONSTRAINT IF EXISTS fk_posts_characters_post;
		ALTER TABLE posts_characters ADD CONSTRAINT fk_posts_characters_post
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE;
		ALTER TABLE posts_characters DROP CONSTRAINT IF EXISTS fk_posts_characters_character;
		ALTER TABLE posts_characters ADD CONSTRAINT fk_posts_characters_character
			FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE;
	`).Error; err != nil {
		return fmt.Errorf("failed to set cascade on posts_characters: %w", err)
	}

	// Prevent duplicate (post, character) pairs.
	if err := db.Exec(
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_characters_post_character ON posts_characters(post_id, character_id)`,
	).Error; err != nil {
		return fmt.Errorf("failed to create posts_characters unique index: %w", err)
	}

	// Speed up category-name and tag-name lookups (case-insensitive search).
	if err := db.Exec(
		`CREATE INDEX IF NOT EXISTS idx_categories_name_lower ON categories(LOWER(name))`,
	).Error; err != nil {
		return fmt.Errorf("failed to create categories name index: %w", err)
	}
	if err := db.Exec(
		`CREATE INDEX IF NOT EXISTS idx_tags_name_lower ON tags(LOWER(name))`,
	).Error; err != nil {
		return fmt.Errorf("failed to create tags name index: %w", err)
	}

	log.Info("Database indexes created successfully")
	return nil
}
