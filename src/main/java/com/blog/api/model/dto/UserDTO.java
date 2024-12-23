package com.blog.api.model.dto;

import jakarta.validation.constraints.Max;
import lombok.Data;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

@Data
public class UserDTO {
    @NotBlank(message = "First name is required")
    @Max(50)
    private String firstName;

    @NotBlank(message = "Username name is required")
    @Max(50)
    private String username;

    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    private String email;

    private String image;
}
