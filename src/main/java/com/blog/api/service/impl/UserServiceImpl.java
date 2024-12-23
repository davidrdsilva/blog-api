package com.blog.api.service.impl;

import com.blog.api.exception.UserNotFoundException;
import com.blog.api.model.dto.UserDTO;
import com.blog.api.model.entity.User;
import com.blog.api.repository.UserRepository;
import com.blog.api.service.UserService;
import lombok.RequiredArgsConstructor;
import org.springframework.dao.DuplicateKeyException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;

    @Override
    @Transactional
    public User createUser(UserDTO userDTO) {
        if (userRepository.existsByEmail(userDTO.getEmail())) {
            throw new DuplicateKeyException("Email already exists: " + userDTO.getEmail());
        }

        User user = new User();
        mapDTOToUser(userDTO, user);

        return userRepository.save(user);
    }

    @Override
    @Transactional
    public User updateUser(UUID id, UserDTO userDTO) {
        User user = getUserById(id);

        // Check if new email is already used by another user
        if (!user.getEmail().equals(userDTO.getEmail()) &&
                userRepository.existsByEmail(userDTO.getEmail())) {
            throw new DuplicateKeyException("Email already exists: " + userDTO.getEmail());
        }

        mapDTOToUser(userDTO, user);
        return userRepository.save(user);
    }

    @Override
    @Transactional
    public void deleteUser(UUID id) {
        if (!userRepository.existsById(id)) {
            throw new UserNotFoundException(id);
        }
        userRepository.deleteById(id);
    }

    @Override
    @Transactional(readOnly = true)
    public User getUserById(UUID id) {
        return userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id));
    }

    @Override
    @Transactional(readOnly = true)
    public List<User> getAllUsers() {
        return userRepository.findAll();
    }

    private void mapDTOToUser(UserDTO dto, User user) {
        user.setFirstName(dto.getFirstName());
        user.setUsername(dto.getUsername());
        user.setEmail(dto.getEmail());
        user.setImage(dto.getImage());
    }
}
