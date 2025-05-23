package com.blog.api.controller;

import com.blog.api.model.dto.CreatePostDTO;
import com.blog.api.model.dto.PostDTO;
import com.blog.api.model.entity.Post;
import com.blog.api.service.PostService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/posts")
@RequiredArgsConstructor
@Tag(name = "Post Management", description = "Routes for managing posts")
public class PostController {

    @Autowired
    private final PostService postService;

    @Operation(
            summary = "Create a new post",
            description = "Inserts a new post with specified values into database"
    )
    @ApiResponse(
            responseCode = "201",
            description = "Successfully created post"
    )
    @PostMapping
    public ResponseEntity<Post> createPost(@RequestBody @Valid CreatePostDTO postDTO) {
        return ResponseEntity.ok(postService.createPost(postDTO));
    }

    @Operation(
            summary = "Get all posts",
            description = "Retrieves all posts"
    )
    @ApiResponse(
            responseCode = "200",
            description = "Successfully retrieved the posts"
    )
    @ApiResponse(
            responseCode = "404",
            description = "Posts were not found"
    )
    @GetMapping()
    public ResponseEntity<List<PostDTO>> getAllPosts() {
        List<Post> posts = postService.getAllPosts();
        List<PostDTO> postDTOs = posts.stream()
                .map(PostDTO::fromEntity)
                .collect(Collectors.toList());
        return ResponseEntity.ok(postDTOs);
    }

    @Operation(
            summary = "Get post by ID",
            description = "Retrieves a post based on its ID"
    )
    @ApiResponse(
            responseCode = "200",
            description = "Successfully retrieved post"
    )
    @ApiResponse(
            responseCode = "404",
            description = "Post not found"
    )
    @GetMapping("/{id}")
    public ResponseEntity<PostDTO> getPost(@PathVariable UUID id) {
        Post post = postService.getPostById(id);
        return ResponseEntity.ok(PostDTO.fromEntity(post));
    }
}
