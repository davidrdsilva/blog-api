package com.blog.api.service;

import com.blog.api.model.dto.CreatePostDTO;
import com.blog.api.model.dto.UpdatePostDTO;
import com.blog.api.model.entity.Post;

import java.util.List;
import java.util.UUID;

public interface PostService {
    Post createPost(CreatePostDTO createPostDTO);
    Post updatePost(UUID id, UpdatePostDTO updatePostDTO);
    void deletePost(UUID id);
    Post getPostById(UUID id);
    List<Post> getAllPosts();
}
