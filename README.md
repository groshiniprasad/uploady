# Receipt Management API

## Overview

The **Receipt Management API** is a backend service that allows users to upload receipt images, store their information, retrieve them, and resize them on demand. The API includes user authentication with JWT tokens to ensure that only authorized users can access their own receipts.

## Features

- **User Authentication**: Secure login and registration with JWT-based authentication.
- **Receipt Upload**: Users can upload receipt images along with metadata (name, amount, date, description).
- **Receipt Retrieval**: Retrieve receipt details by ID, including an option to resize the image dynamically.
- **Image Resizing**: Receipts can be fetched as resized images with width and height parameters.
- **Validation**: Input validation for file size, file type, and metadata.

## Technologies

- **Language**: Go (Golang)
- **Database**: MySQL
- **Authentication**: JWT (JSON Web Tokens)
- **Image Processing**: "golang.org/x/image/draw"
- **Router**: `github.com/gorilla/mux`

## Installation

1. **Clone the Repository**:

   git clone [uploady rep](https://github.com/groshiniprasad/uploady.git)
   cd uploady

2. **Set Up Environment file**:
    PORT=
    DB_USER=
    DB_PASSWORD=
    DB_HOST=
    DB_PORT=
    DB_NAME=
    JWT_SECRET=
    JWTExpirationInSeconds=
## 3. Set Up Database

To set up the database and run the application, follow these steps:

1. **Navigate to the project directory**:
   ```bash
   cd uploady
    Ensure the Makefile is present in the current directory: Check by running:
    ls
2. **Set Up the databse**:
Run the following commands to set up the database and start the server:
   ```bash
   1. make create-database
   2. make migrate-up
   3. make run
    