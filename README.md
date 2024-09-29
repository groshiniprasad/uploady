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

   git clone [https://github.com/yourusername/receipt-management-api.git](https://github.com/groshiniprasad/uploady.git)
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
   1. make create-database
   2. make migrate-up
   3. make run

1. **Testing APIS**:
After setting up the database and running the server, you can test the API endpoints.

1. Register Users
To register a user, open a new terminal window (make sure the server is running in the other terminal) and use the following curl command:

curl -X POST http://localhost:8080/api/v1/register \
-d '{"FirstName": "RP", "LastName": "Se", "email": "rp@side.com", "password": "password123"}' \
-H "Content-Type: application/json"

2. Login Users
    To log in and get a JWT token, use the following command:
    curl -X POST http://localhost:8080/api/v1/login \
-d '{"email": "rp@side.com", "password": "password123"}' \
-H "Content-Type: application/json"

3. Upload a Receipt Image
To upload a receipt image along with metadata, use the JWT token obtained from the login step. Replace <your_token> with the actual token in the command below:
curl -X POST http://localhost:8080/api/v1/receipts/upload \
-H "Authorization: Bearer <your_token>" \
-F "Name=Grocery Store Receipt" \
-F "amount=56.99" \
-F "date=2024-10-01" \
-F "image=@/Users/rpg/Documents/Pictures/RPG.jpg"

4. Resize and Download the Image
To resize and download the image, use the following command. Make sure to replace <your_token> with your actual JWT token:
curl "http://localhost:8080/api/v1/receipts/1?width=200&height=200" \
-H "Authorization: Bearer <your_token>" \
--output resized_image.jpg && open resized_image.jpg

Notes:
Make sure your database and server are running before testing any of the APIs.
Replace <your_token> in the Authorization header with the actual JWT token obtained during the login process.
You can modify the width and height parameters in the resize request to get different dimensions for the resized image.