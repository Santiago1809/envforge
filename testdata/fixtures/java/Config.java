package com.example.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;

@Configuration
public class DatabaseConfig {
    
    @Value("${DB_HOST}")
    private String host;
    
    @Value("${DB_PORT}")
    private int port;
    
    @Value("${API_KEY}")
    private String apiKey;
    
    public String getHost() {
        return System.getenv("DB_HOST");
    }
    
    public int getPort() {
        return Integer.parseInt(System.getProperty("DB_PORT", "5432"));
    }
}
