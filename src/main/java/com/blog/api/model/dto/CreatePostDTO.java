package com.blog.api.model.dto;

import com.fasterxml.jackson.annotation.JsonRawValue;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import lombok.Data;

import java.util.UUID;

@Data
public class CreatePostDTO {
    @NotBlank
    @Size(max = 100)
    private String title;

    @NotBlank
    @Size(max = 255)
    private String description;

    @Size(max = 500)
    private String image;

    @NotBlank
    private UUID authorId;

    @NotNull
    @JsonRawValue
    private Object body;  // JSON string
}
