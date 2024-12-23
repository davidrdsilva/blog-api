package com.blog.api.model.dto;

import jakarta.validation.constraints.Size;
import lombok.Data;

@Data
public class UpdatePostDTO {
    @Size(max = 100)
    private String title;

    @Size(max = 255)
    private String description;

    @Size(max = 500)
    private String image;

    private String body;  // JSON string
}
