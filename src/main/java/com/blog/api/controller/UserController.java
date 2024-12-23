package com.blog.api.controller;

import com.blog.api.model.dto.UserDTO;
import com.blog.api.model.entity.User;
import com.blog.api.service.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;

@RestController
@RequestMapping("/api/users")
@RequiredArgsConstructor
@Tag(name = "User Management", description = "APIs for managing users")
public class UserController {

    @Autowired
    private final UserService userService;

    @Operation(
            summary = "Create a new user",
            description = "Inserts a new user with specified values into database"
    )
    @ApiResponse(
            responseCode = "201",
            description = "Successfully created user"
    )
    @PostMapping
    public ResponseEntity<User> createUser(@RequestBody @Valid UserDTO userDTO) {
        return ResponseEntity.ok(userService.createUser(userDTO));
    }

    @Operation(
            summary = "Get user by ID",
            description = "Retrieves a user based on their ID"
    )
    @ApiResponse(
            responseCode = "200",
            description = "Successfully retrieved user"
    )
    @ApiResponse(
            responseCode = "404",
            description = "User not found"
    )
    @GetMapping("/{id}")
    public ResponseEntity<User> getUser(@PathVariable UUID id) {
        return ResponseEntity.ok(userService.getUserById(id));
    }
}
