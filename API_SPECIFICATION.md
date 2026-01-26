# Blog API Specification

This document defines the REST API endpoints required by the blog-client frontend application.
You must write a REST API using this stack:
- Go
- Gin, for the web framework
- GORM, for the database (PostgreSQL)
- Docker, for containerization
- MinIO, for object storage

### Project standards

- The project must follow the hexagonal architecture
- The project must be containerized and dockerized
- The code should be clean, modularized and loosely coupled
- The comments shoudl focus on why instead of how
- There must be rich, configurable logs separated by color, for example:

```bash
[2024-01-15 10:30:00] [INFO] [POST /api/posts] Post created successfully
```

## Base URL

```
/api/v1
```

## Content Type

All endpoints accept and return `application/json` unless otherwise specified.

## Authentication

Authentication mechanism is not currently implemented in the frontend. Future implementations should use Bearer tokens in the `Authorization` header.

---

## Data Models

### Post

```typescript
interface Post {
    id: string;                    // UUID v4
    title: string;                 // Required, 1-200 characters
    subtitle: string | null;       // Optional, max 300 characters
    description: string;           // Required, 1-100 characters
    image: string;                 // Required, valid URL
    date: string;                  // ISO 8601 datetime
    author: string;                // Required, 1-100 characters
    content: EditorJsContent | null;
    createdAt: string;             // ISO 8601 datetime
    updatedAt: string;             // ISO 8601 datetime
}
```

### EditorJsContent

```typescript
interface EditorJsContent {
    blocks: EditorJsBlock[];
    time: number;                  // Unix timestamp in milliseconds
    version: string;               // Editor.js version (e.g., "2.28.0")
}
```

### EditorJsBlock

```typescript
interface EditorJsBlock {
    id: string;                    // Block identifier
    type: string;                  // Block type (paragraph, header, list, etc.)
    data: Record<string, unknown>; // Block-specific data
}
```

Supported block types:
- `paragraph`: `{ text: string }`
- `header`: `{ text: string, level: 2 | 3 | 4 }`
- `list`: `{ style: "ordered" | "unordered", items: string[] }`
- `quote`: `{ text: string, caption?: string }`
- `code`: `{ code: string }`
- `image`: `{ file?: { url: string }, url?: string, caption?: string }`
- `linkTool`: `{ link: string, meta?: { title?: string, description?: string, image?: { url: string } } }`

### Error Response

```typescript
interface ErrorResponse {
    error: {
        code: string;              // Machine-readable error code
        message: string;           // Human-readable error message
        details?: Record<string, string[]>;  // Field-level validation errors
    };
}
```

### Pagination Metadata

```typescript
interface PaginationMeta {
    total: number;                 // Total number of items
    page: number;                  // Current page (1-indexed)
    limit: number;                 // Items per page
    totalPages: number;            // Total number of pages
    hasMore: boolean;              // Whether more pages exist
}
```

---

## Endpoints

### Posts

#### List Posts

Retrieves a paginated list of posts with optional search and filtering.

```
GET /api/posts
```

**Query Parameters**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-indexed) |
| `limit` | integer | 6 | Items per page (max: 50) |
| `search` | string | - | Search query (searches title, subtitle, description, content) |
| `author` | string | - | Filter by author name |
| `sortBy` | string | "date" | Sort field: "date", "title", "createdAt", "updatedAt" |
| `sortOrder` | string | "desc" | Sort order: "asc", "desc" |

**Response**

```
200 OK
```

```json
{
    "data": [
        {
            "id": "550e8400-e29b-41d4-a716-446655440000",
            "title": "The Nature of Consciousness",
            "subtitle": "Exploring what it means to be aware",
            "description": "A deep dive into consciousness and awareness.",
            "image": "https://images.unsplash.com/photo-1506905925346-21bda4d32df4",
            "date": "2024-01-15T00:00:00.000Z",
            "author": "David",
            "content": null,
            "createdAt": "2024-01-15T10:30:00.000Z",
            "updatedAt": "2024-01-15T10:30:00.000Z"
        }
    ],
    "meta": {
        "total": 12,
        "page": 1,
        "limit": 6,
        "totalPages": 2,
        "hasMore": true
    }
}
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `INVALID_QUERY_PARAM` | Invalid query parameter value |

---

#### Get Post

Retrieves a single post by ID with full content.

```
GET /api/posts/:id
```

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Post UUID |

**Response**

```
200 OK
```

```json
{
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "The Nature of Consciousness",
        "subtitle": "Exploring what it means to be aware",
        "description": "A deep dive into consciousness and awareness.",
        "image": "https://images.unsplash.com/photo-1506905925346-21bda4d32df4",
        "date": "2024-01-15T00:00:00.000Z",
        "author": "David",
        "content": {
            "blocks": [
                {
                    "id": "block-1",
                    "type": "paragraph",
                    "data": {
                        "text": "Consciousness remains one of the most profound mysteries."
                    }
                },
                {
                    "id": "block-2",
                    "type": "header",
                    "data": {
                        "text": "The Hard Problem",
                        "level": 2
                    }
                }
            ],
            "time": 1705276800000,
            "version": "2.28.0"
        },
        "createdAt": "2024-01-15T10:30:00.000Z",
        "updatedAt": "2024-01-15T10:30:00.000Z"
    }
}
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 404 | `POST_NOT_FOUND` | Post with specified ID does not exist |
| 400 | `INVALID_POST_ID` | Invalid UUID format |

---

#### Create Post

Creates a new blog post.

```
POST /api/posts
```

**Workflow**

The featured image must be uploaded separately before creating the post:
1. Upload the image via `POST /api/upload`
2. Receive the image URL in the upload response
3. Include the image URL in the create post request

**Request Body**

```json
{
    "title": "My New Post",
    "subtitle": "An optional subtitle",
    "description": "A short description for the post card.",
    "image": "https://storage.example.com/uploads/featured-123456.jpg",
    "author": "David",
    "content": {
        "blocks": [
            {
                "id": "block-1",
                "type": "paragraph",
                "data": {
                    "text": "This is the first paragraph."
                }
            }
        ],
        "time": 1705276800000,
        "version": "2.28.0"
    }
}
```

**Validation Rules**

| Field | Rules |
|-------|-------|
| `title` | Required, string, 1-200 characters |
| `subtitle` | Optional, string, max 300 characters |
| `description` | Required, string, 1-100 characters |
| `image` | Required, valid URL (must be from trusted storage domain) |
| `author` | Required, string, 1-100 characters |
| `content` | Optional, valid EditorJsContent object |

**Image URL Validation**

The `image` field should only accept URLs from the application's file storage:
- Must be a valid URL format
- Should match the storage domain configured for uploads
- Recommended: Validate that the file exists in storage before accepting

**Response**

```
201 Created
```

```json
{
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "title": "My New Post",
        "subtitle": "An optional subtitle",
        "description": "A short description for the post card.",
        "image": "https://storage.example.com/uploads/featured-123456.jpg",
        "date": "2024-01-20T14:30:00.000Z",
        "author": "David",
        "content": { ... },
        "createdAt": "2024-01-20T14:30:00.000Z",
        "updatedAt": "2024-01-20T14:30:00.000Z"
    }
}
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `VALIDATION_ERROR` | Request body validation failed |
| 400 | `INVALID_CONTENT_FORMAT` | EditorJs content structure is invalid |
| 400 | `INVALID_IMAGE_URL` | Image URL is not from trusted storage |

Example validation error:

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Request validation failed",
        "details": {
            "title": ["Title is required", "Title must not exceed 200 characters"],
            "description": ["Description must not exceed 100 characters"],
            "image": ["Image is required", "Image must be uploaded via /api/upload"]
        }
    }
}
```

---

#### Update Post

Updates an existing blog post.

```
PUT /api/posts/:id
```

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Post UUID |

**Workflow**

To update the featured image:
1. Upload the new image via `POST /api/upload`
2. Receive the new image URL in the upload response
3. Include the new image URL in the update request

**Request Body**

All fields are optional. Only provided fields will be updated.

```json
{
    "title": "Updated Post Title",
    "subtitle": "Updated subtitle",
    "description": "Updated description.",
    "image": "https://storage.example.com/uploads/new-featured-789.jpg",
    "content": {
        "blocks": [ ... ],
        "time": 1705363200000,
        "version": "2.28.0"
    }
}
```

**Validation Rules**

Same as Create Post, but all fields are optional. When `image` is provided, it must be a valid URL from trusted storage.

**Response**

```
200 OK
```

```json
{
    "data": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "title": "Updated Post Title",
        "subtitle": "Updated subtitle",
        "description": "Updated description.",
        "image": "https://storage.example.com/uploads/new-featured-789.jpg",
        "date": "2024-01-15T00:00:00.000Z",
        "author": "David",
        "content": { ... },
        "createdAt": "2024-01-15T10:30:00.000Z",
        "updatedAt": "2024-01-21T09:15:00.000Z"
    }
}
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `VALIDATION_ERROR` | Request body validation failed |
| 400 | `INVALID_POST_ID` | Invalid UUID format |
| 400 | `INVALID_IMAGE_URL` | Image URL is not from trusted storage |
| 404 | `POST_NOT_FOUND` | Post with specified ID does not exist |

---

#### Delete Post

Permanently deletes a blog post.

```
DELETE /api/posts/:id
```

**Path Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Post UUID |

**Response**

```
204 No Content
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `INVALID_POST_ID` | Invalid UUID format |
| 404 | `POST_NOT_FOUND` | Post with specified ID does not exist |

---

### File Upload

#### Upload Image

Uploads an image file for use in Editor.js image blocks.

```
POST /api/upload
```

**Request**

Content-Type: `multipart/form-data`

| Field | Type | Description |
|-------|------|-------------|
| `file` | File | Image file (JPEG, PNG, GIF, WebP) |

**Validation Rules**

| Rule | Value |
|------|-------|
| Max file size | 5 MB |
| Allowed MIME types | image/jpeg, image/png, image/gif, image/webp |
| Max dimensions | 4096x4096 pixels |

**Response**

```
200 OK
```

Editor.js Image Tool expects this specific response format:

```json
{
    "success": 1,
    "file": {
        "url": "https://storage.example.com/uploads/image-123456.jpg"
    }
}
```

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `NO_FILE_PROVIDED` | No file in request |
| 400 | `INVALID_FILE_TYPE` | File type not allowed |
| 400 | `FILE_TOO_LARGE` | File exceeds size limit |
| 500 | `UPLOAD_FAILED` | Server failed to process upload |

Error response format for Editor.js:

```json
{
    "success": 0,
    "error": {
        "code": "FILE_TOO_LARGE",
        "message": "File size exceeds the 5MB limit"
    }
}
```

---

### URL Metadata

#### Fetch URL Metadata

Fetches metadata (title, description, image) from a URL for Editor.js Link Tool.

```
GET /api/fetch-url
```

**Query Parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `url` | string | URL to fetch metadata from (required) |

**Response**

```
200 OK
```

Editor.js Link Tool expects this specific response format:

```json
{
    "success": 1,
    "link": "https://example.com/article",
    "meta": {
        "title": "Example Article Title",
        "description": "A brief description of the article content.",
        "image": {
            "url": "https://example.com/og-image.jpg"
        }
    }
}
```

**Metadata Extraction Priority**

1. `title`: Open Graph `og:title` > `<title>` tag
2. `description`: Open Graph `og:description` > `<meta name="description">`
3. `image.url`: Open Graph `og:image` > Twitter `twitter:image`

**Error Responses**

| Status | Code | Description |
|--------|------|-------------|
| 400 | `INVALID_URL` | URL parameter is missing or malformed |
| 400 | `URL_NOT_ACCESSIBLE` | Unable to fetch the URL |
| 408 | `REQUEST_TIMEOUT` | URL fetch timed out (>10 seconds) |

Error response format for Editor.js:

```json
{
    "success": 0,
    "error": {
        "code": "URL_NOT_ACCESSIBLE",
        "message": "Unable to fetch metadata from the provided URL"
    }
}
```

---

## HTTP Status Codes Summary

| Status | Description |
|--------|-------------|
| 200 | Successful request |
| 201 | Resource created successfully |
| 204 | Successful request with no content (delete) |
| 400 | Bad request (validation error, invalid parameters) |
| 404 | Resource not found |
| 408 | Request timeout |
| 500 | Internal server error |

---

## Rate Limiting

Recommended rate limits:

| Endpoint Category | Limit |
|------------------|-------|
| Read operations (GET) | 100 requests/minute |
| Write operations (POST, PUT, DELETE) | 20 requests/minute |
| File uploads | 10 requests/minute |

Rate limit headers to include in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705363200
```

---

## CORS Configuration

The API should allow requests from the frontend origin:

```
Access-Control-Allow-Origin: <frontend-origin>
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

---

## Search Implementation Notes

The search functionality should support:

1. **Full-text search** across:
   - Post title
   - Post subtitle
   - Post description
   - Content blocks (extract text from paragraph, header, quote, code, list items)

2. **Multi-term search**: All search terms must match (AND logic)

3. **Case-insensitive matching**

4. **Recommended indexing**:
   - Create a computed `searchable_text` column combining all searchable fields
   - Use database full-text search (PostgreSQL `tsvector`, MySQL `FULLTEXT`, etc.)

---

## Database Schema Suggestion

```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    subtitle VARCHAR(300),
    description VARCHAR(100) NOT NULL,
    image VARCHAR(2048) NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    author VARCHAR(100) NOT NULL,
    content JSONB,
    searchable_text TEXT GENERATED ALWAYS AS (
        title || ' ' || COALESCE(subtitle, '') || ' ' || description
    ) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_date ON posts(date DESC);
CREATE INDEX idx_posts_author ON posts(author);
CREATE INDEX idx_posts_searchable_text ON posts USING GIN(to_tsvector('english', searchable_text));
```

---

## Implementation Checklist

- [ ] `GET /api/posts` - List posts with pagination and search
- [ ] `GET /api/posts/:id` - Get single post
- [ ] `POST /api/posts` - Create post
- [ ] `PUT /api/posts/:id` - Update post
- [ ] `DELETE /api/posts/:id` - Delete post
- [ ] `POST /api/upload` - Upload image (Editor.js format)
- [ ] `GET /api/fetch-url` - Fetch URL metadata (Editor.js format)
- [ ] Input validation for all endpoints
- [ ] Error response standardization
- [ ] CORS configuration
- [ ] Rate limiting (optional)
