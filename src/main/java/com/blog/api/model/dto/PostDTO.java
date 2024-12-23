package com.blog.api.model.dto;

import com.blog.api.model.entity.Post;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
public class PostDTO {
    private UUID id;
    private String title;
    private String description;
    private String image;
    private Integer views;
    private String body;  // JSON string
    private UUID authorId;  // Instead of full User object, just the ID
    private String authorName;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    // Constructor to convert from Entity to DTO
    public static PostDTO fromEntity(Post post) {
        PostDTO dto = new PostDTO();
        dto.setId(post.getId());
        dto.setTitle(post.getTitle());
        dto.setDescription(post.getDescription());
        dto.setImage(post.getImage());
        dto.setViews(post.getViews());
        dto.setBody(post.getBody());
        dto.setAuthorId(post.getUser().getId());  // Assuming User has getId()
        dto.setAuthorName(post.getUser().getFirstName());
        dto.setCreatedAt(post.getCreatedAt());
        dto.setUpdatedAt(post.getUpdatedAt());
        return dto;
    }
}
