package com.blog.api.service.impl;

import com.blog.api.exception.UserNotFoundException;
import com.blog.api.model.dto.CreatePostDTO;
import com.blog.api.model.dto.UpdatePostDTO;
import com.blog.api.model.entity.Post;
import com.blog.api.model.entity.User;
import com.blog.api.repository.PostRepository;
import com.blog.api.repository.UserRepository;
import com.blog.api.service.PostService;
import lombok.RequiredArgsConstructor;
import org.springframework.dao.DuplicateKeyException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class PostServiceImpl implements PostService {

    private final PostRepository postRepository;
    private final UserRepository userRepository;

    @Override
    @Transactional
    public Post createPost(CreatePostDTO createPostDTO) {
        if (postRepository.existsByTitle(createPostDTO.getTitle())) {
            throw new DuplicateKeyException("This post already exists: " + createPostDTO.getTitle());
        }

        Post post = new Post();

        post.setTitle(createPostDTO.getTitle());
        post.setDescription(createPostDTO.getDescription());
        post.setImage(createPostDTO.getImage());
        post.setBody(createPostDTO.getBody());

        if (!userRepository.existsById(createPostDTO.getAuthorId())) {
            throw new UserNotFoundException(createPostDTO.getAuthorId());
        }

        User user = userRepository.findById(createPostDTO.getAuthorId()).orElseThrow(() -> new UserNotFoundException(createPostDTO.getAuthorId()));

        post.setUser(user);

        return postRepository.save(post);
    }

    @Override
    public Post updatePost(UUID id, UpdatePostDTO updatePostDTO) {
        return null;
    }

    @Override
    public void deletePost(UUID id) {

    }

    @Override
    public Post getPostById(UUID id) {
        return postRepository.getReferenceById(id);
    }

    @Override
    public List<Post> getAllPosts() {
        return List.of();
    }
}
