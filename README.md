# Full-Stack Todo List Application

[Backend API Documentation](https://todolist.apidocumentation.com/)

## 1. Overview

This report covers the design, tech stack, and setup process for our Todo List app. The app lets users manage tasks, add tags, create subtasks, and attach images. Users can sign in with email/password or Google. We built the backend with Go using Hexagonal Architecture and the frontend with Next.js, React, and TypeScript.

## 2. Tech Stack

### Backend:

* **Language:** Go
* **Web Framework:** Chi (v5) - Handles HTTP routing and middleware
* **Database Tools:**
  * `pgx/v5`: PostgreSQL driver
  * `sqlc`: Turns SQL queries into Go code
  * `golang-migrate`: Manages database schema changes
* **API Tools:**
  * OpenAPI 3.0: Defines the API
  * `oapi-codegen`: Creates Go code from OpenAPI specs
* **Auth Tools:**
  * JWT: For token generation and checking
  * bcrypt: For password hashing
  * OAuth2: For Google login
* **Other Tools:**
  * `viper`: Manages app settings
  * `slog`: For logging
  * `go-cache`: For temporary data storage
  * Google Cloud Storage: For file storage

### Frontend:

* **Framework:** Next.js (v15+) with App Router
* **Language:** TypeScript
* **UI:** React (v19+)
* **Styling:** Tailwind CSS (v4+)
* **Component Library:** shadcn/ui (built on Radix UI)
* **State Management:**
  * Zustand: For app-wide state (like login status)
  * TanStack Query: For server data, caching, and syncing
* **Forms:** React Hook Form
* **Drag & Drop:** @dnd-kit (for Kanban board)
* **Other Tools:** sonner (for notifications), date-fns, lucide-react (icons)

## 3. Database

* **Type:** PostgreSQL
* **Tables:**
  * `users`: Stores user info and login details
  * `tags`: User-created labels with name, color, and icon
  * `todos`: Main task items with title, description, status, deadline, and attachment link
  * `subtasks`: Step-by-step items for each todo
  * `todo_tags`: Links todos to tags (many-to-many)
* **Features:**
  * Indexes on frequently searched columns for speed
  * Auto-updated timestamp columns

## 4. Setup Guide

### Requirements:

* Go (version 1.24+)
* Node.js (v20+ recommended)
* pnpm (or npm/yarn)
* PostgreSQL Database
* Command line tools: `migrate`, `sqlc`, `oapi-codegen`
* Google Cloud Storage account and bucket

### Backend Setup:

1. **Get the code:** `git clone https://github.com/Sosokker/go-chi-oapi-codegen-todolist.git`
2. **Go to backend folder:** `cd backend`
3. **Set up config:**
   * Copy `example.config.yaml` to `config.yaml`
   * Update database connection, JWT secret, Google OAuth settings, and GCS bucket name
   * Add your GCS credentials file
4. **Set up database:**
   * Use your own PostgreSQL server and provide postgres connection string
5. **Update `DB_URL` and `MIGRATIONS_PATH`** in `Makefile`
   * `DB_URL`: Your PostgreSQL connection string
   * `MIGRATIONS_PATH`: Path to your migrations directory (Absolute Path)
6. **Run database setup:** `make migrate-up`
7. **Start the backend:**
   * Regular mode: `make run`
   * Development mode: `make dev` (auto-reloads on changes)

### Frontend Setup:

1. **Go to frontend folder:** `cd ../frontend`
2. **Install packages:** `pnpm install`
3. **Create config file:**
   * Make a `.env.local` file with:
   ```
   NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
   ```
4. **Start the frontend:** `pnpm dev`

The app should now run at `http://localhost:3000`.

## 5. Code Architecture

### Overall Design:

* **Backend:** Provides a REST API, handles business logic, data storage, and auth
* **Frontend:** Creates user interface and manages user interactions
* **Communication:** Frontend sends HTTP requests to backend API endpoints

### Backend Architecture - Hexagonal (Ports & Adapters):

The backend follows a "ports and adapters" pattern for better separation of concerns:

1. **Core Domain:** Basic business objects (`User`, `Todo`, `Tag`) and logic
2. **Service Layer:** Contains the main business rules and coordinates operations
3. **Ports (Interfaces):** Define contracts for different parts of the system
4. **Adapters:** Implement the interfaces, connecting the app to external systems:
   * API handlers (connect HTTP to services)
   * Database repositories (connect services to database)
   * Storage service (connects to Google Cloud)

The goal is to **decouple core business logic** from external systems such as databases, web frameworks, and third-party APIs.

#### Core (Inside the Hexagon)

- Contains essential **business rules and domain logic**
- Located in:
  - `internal/domain`: Domain models and error definitions
  - `internal/service`: Use cases and service logic
- The core is **technology-agnostic** and has **no knowledge** of infrastructure or frameworks

#### Ports (Hexagon Boundary Interfaces)

Interfaces that define how the core interacts with the outside world.

- **Driving Ports** (Input Interfaces)  
  - Describe how the application can be used  
  - Example: `TodoService` interface defines operations like `CreateTodo`, `ListTodos`  
  - Implemented in the **application layer** (`internal/service`)

- **Driven Ports** (Output Interfaces)  
  - Describe how the application depends on external systems  
  - Example: `TodoRepository`, `FileStorageService` interfaces  
  - Defined in service or repository packages as needed

#### Adapters (Outside the Hexagon)

Concrete implementations that **connect ports to external systems**.

- **Driving Adapters**  
  - Translate external input (HTTP, CLI, etc.) into calls to the core via driving ports  
  - Example:  
    - HTTP handlers in `internal/api` invoke `TodoService` methods  
    - Middleware handles auth and CORS

- **Driven Adapters**  
  - Provide infrastructure-specific implementations of driven ports  
  - Examples:
    - `pgxTodoRepository` in `internal/repository` implements `TodoRepository` using PostgreSQL
    - `gcsStorageService` in `internal/service` implements `FileStorageService` using GCS

### Frontend Architecture:

1. **Pages & Routes:** In the `app/` folder using Next.js App Router
2. **Components:**
   * `components/ui/`: Basic UI elements (buttons, inputs, etc.)
   * `components/`: App-specific components (TodoCard, TodoForm, etc.)
3. **State Management:**
   * Server data: TanStack Query fetches and caches API data
   * Client state: Zustand manages user login status
4. **API Calls:** Centralized in service files for consistency

### Key Design Choices:

* **Hexagonal Architecture:** Separates business logic from external systems
* **OpenAPI First:** Clearly defines API before coding
* **sqlc:** Generates type-safe database code
* **Cloud Storage:** Stores file attachments separately from database
* **Caching:** Improves performance for frequently accessed data
* **Modern Frontend:** Uses latest React patterns and tools
