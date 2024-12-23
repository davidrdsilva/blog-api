package com.blog.api.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class SwaggerConfig {

    @Bean
    public OpenAPI springOpenAPI() {
        return new OpenAPI()
                .info(new Info().title("Blog API")
                        .description("This is the official documentation of the Blog's main API")
                        .version("v1.0.0")
                        .contact(new Contact().name("David").email("david@blog.com").url("https://www.blog.com"))
                        .license(new License().name("MIT").url("https://opensource.org/license/mit")));
    }
}
