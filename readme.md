# Movie Reservation System 🎬

A full-stack backend application for booking movie tickets built with **Go**, **Gin**, and **GORM**. This project is part of the [Roadmap.sh Backend Projects](https://roadmap.sh/projects).

## Features

### User Management
- User registration and login with JWT authentication
- Role-based access control (Admin & Regular User)
- Password hashing using bcrypt

### Movie Management (Admin Only)
- Add, view movies
- Support for multiple genres per movie
- Movie details with poster, duration, and release date

### Showtime Management (Admin Only)
- Create showtimes for movies
- Automatic end time calculation based on movie duration
- Available seats tracking

### Reservation System
- Reserve multiple seats for a showtime
- Real-time seat availability checking
- Prevent overbooking with database transactions and locking
- Cancel upcoming reservations only

### Technical Features
- Secure JWT authentication
- Database transactions for data consistency
- Pessimistic locking to prevent race conditions
- Soft delete support
- Proper error handling and validation

## Tech Stack

- **Language**: Go (Golang)
- **Framework**: Gin Gonic
- **ORM**: GORM v2
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Password Hashing**: bcrypt

## Project Structure


