package com.blog.api.repository;

import com.blog.api.model.entity.Post;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.UUID;

public interface PostRepository extends JpaRepository<Post, UUID> {
    boolean existsByTitle(String title);
}
