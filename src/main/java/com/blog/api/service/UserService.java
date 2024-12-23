package com.blog.api.service;

import com.blog.api.model.dto.UserDTO;
import com.blog.api.model.entity.User;

import java.util.List;
import java.util.UUID;

public interface UserService {
    User createUser(UserDTO userDTO);
    User updateUser(UUID id, UserDTO userDTO);
    void deleteUser(UUID id);
    User getUserById(UUID id);
    List<User> getAllUsers();
}
