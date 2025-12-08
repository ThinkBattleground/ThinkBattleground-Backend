# Architecture & Data Flow

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT APPLICATION                        │
│                    (Frontend / Web / Mobile)                     │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     │ HTTP Request
                     │ + Firebase Token
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GIN WEB SERVER (Port 8080)                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────┐         ┌──────────────────┐              │
│  │ Routes Handler   │         │ Auth Middleware  │              │
│  │ /api/v1/...      │◄────────┤ Verify Token     │              │
│  └────────┬─────────┘         └──────────────────┘              │
│           │                                                       │
│           │                   ┌──────────────────┐              │
│           │                   │ Admin Middleware │              │
│           │                   │ Check is_admin   │              │
│           └──────────┬────────┤                  │              │
│                      │        └──────────────────┘              │
│  ┌────────────────────────────┐                                │
│  │  Question Controller       │                                │
│  ├────────────────────────────┤                                │
│  │ • CreateQuestionWithGemini │                                │
│  │ • ListQuestions            │                                │
│  │ • GetQuestion              │                                │
│  │ • GetQuestionsByCategory   │                                │
│  └────────┬───────────────────┘                                │
│           │                                                      │
│  ┌────────────────────────────┐                                │
│  │  Config / Gemini Service   │                                │
│  ├────────────────────────────┤                                │
│  │ • GenerateQuestion()       │                                │
│  │ • InitializeGemini()       │                                │
│  └────────┬───────────────────┘                                │
└───────────┼──────────────────────────────────────────────────────┘
            │
            ├─────────────────────────┐
            │                         │
            ▼                         ▼
    ┌──────────────────┐     ┌──────────────────┐
    │  GEMINI API      │     │  POSTGRESQL      │
    │  (Google)        │     │  DATABASE        │
    │                  │     │                  │
    │ • Generate       │     │ • Store          │
    │ • Parse JSON     │     │ • Index          │
    │ • Return Data    │     │ • Query          │
    └──────────────────┘     └──────────────────┘
```

## Request Flow - Generate Question

```
1. CLIENT REQUEST
   POST /api/v1/admin/questions/generate
   Header: Authorization: Bearer <token>
   Body: {category: "algebra", difficulty: "intermediate"}
                    ▼
2. AUTHENTICATION
   Middleware validates Firebase token
   Extracts user ID and sets in context
                    ▼
3. AUTHORIZATION
   Admin middleware checks is_admin flag
   Returns 403 if not admin
                    ▼
4. VALIDATION
   Controller validates:
   - Required fields present
   - Difficulty is valid
   - Category is valid
                    ▼
5. GEMINI API CALL
   config.GenerateQuestion() sends request:
   - System prompt (guidelines)
   - Human prompt (category + difficulty)
   - API key authentication
   - 30-second timeout
                    ▼
6. JSON PARSING
   Parse Gemini's JSON response
   Extract all question fields
                    ▼
7. MODEL CONVERSION
   parseGeneratedQuestion() converts to Question model
   - Set question_id (UUID if not provided)
   - Set created_by (admin user ID)
   - Validate required fields
                    ▼
8. DATABASE SAVE
   database.DB.Create(&question)
   GORM saves to PostgreSQL
                    ▼
9. RESPONSE
   HTTP 201 Created
   Return saved question object
```

## Data Flow - Retrieve Questions

```
1. CLIENT REQUEST
   GET /api/v1/questions?category=algebra
   Header: Authorization: Bearer <token>
                    ▼
2. AUTHENTICATION
   Validate Firebase token
                    ▼
3. DATABASE QUERY
   database.DB.Find(&questions)
   PostgreSQL returns matching records
                    ▼
4. FILTER (Optional)
   If category parameter:
   database.DB.Where("category = ?", category)
                    ▼
5. RESPONSE
   HTTP 200 OK
   Return array of questions
```

## Database Schema

```
┌──────────────────────────────────────┐
│            questions                 │
├──────────────────────────────────────┤
│ id: BIGSERIAL PRIMARY KEY            │
│ question_id: VARCHAR(36) UNIQUE      │ ◄─ UUID from Gemini
│ title: VARCHAR(255)                  │
│ question: TEXT                       │
│ answer: TEXT                         │
│ explanation: TEXT                    │
│ hints: TEXT[]                        │ ◄─ Array support
│ difficulty: VARCHAR(20)              │
│ expected_time: INTEGER               │
│ points: INTEGER                      │
│ category: VARCHAR(100) [INDEXED]     │
│ sub_category: VARCHAR(100)           │
│ tags: TEXT[]                         │
│ requirements: TEXT[]                 │
│ image_url: TEXT                      │
│ created_by: BIGINT [FOREIGN KEY]     │ ◄─ User ID
│ created_at: TIMESTAMP                │
│ updated_at: TIMESTAMP                │
│ deleted_at: TIMESTAMP [INDEXED]      │
└──────────────────────────────────────┘
            ▲
            │ References
            │
         ┌──────────────────┐
         │     users        │
         │    (id)          │
         └──────────────────┘
```

## Component Interaction Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      routes/routes.go                            │
│  Initializes controllers and defines all endpoints               │
└──────────────────┬──────────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
   ┌─────────────┐   ┌──────────────────┐
   │Puzzle       │   │Question          │
   │Controller   │   │Controller        │
   └─────────────┘   ├──────────────────┤
                     │CreateQuestionWith│
                     │  Gemini()        │
                     │ListQuestions()   │
                     │GetQuestion()     │
                     │GetQuestionsBycat │
                     └────────┬─────────┘
                              │
                              ▼
                      ┌──────────────────┐
                      │config/gemini.go  │
                      ├──────────────────┤
                      │InitializeGemini()│
                      │GenerateQuestion()│
                      └────────┬─────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
                    ▼                     ▼
              ┌────────────┐      ┌──────────────┐
              │models/     │      │database/db.go│
              │question.go │      ├──────────────┤
              ├────────────┤      │InitDB()      │
              │Question{}  │      │AutoMigrate() │
              │Request{}   │      └──────┬───────┘
              └────────────┘             │
                                         ▼
                            ┌────────────────────┐
                            │  PostgreSQL        │
                            │  questions TABLE   │
                            └────────────────────┘
```

## Middleware Chain

```
CLIENT REQUEST
    ▼
/api/v1/admin/questions/generate
    ▼
┌──────────────────────────────────┐
│ Route Defined in routes.go       │
└──────────────────────────────────┘
    ▼
┌──────────────────────────────────┐
│ AuthMiddleware                   │
│ - Extract Firebase token         │
│ - Validate token signature       │
│ - Get user from Firebase         │
│ - Store in context["firebaseUID"]│
└──────────────────────────────────┘
    │
    ├─ Invalid? Return 401
    │
    ▼
┌──────────────────────────────────┐
│ AdminMiddleware                  │
│ - Get user from database         │
│ - Check is_admin flag            │
│ - Store in context["dbUser"]     │
└──────────────────────────────────┘
    │
    ├─ Not admin? Return 403
    │
    ▼
┌──────────────────────────────────┐
│ QuestionController               │
│ .CreateQuestionWithGemini()      │
└──────────────────────────────────┘
    ▼
CLIENT RESPONSE
```

## Error Handling Flow

```
Request arrives
    ▼
Validate JSON binding
    ├─ Invalid? Return 400 Bad Request
    ▼
Validate difficulty level
    ├─ Invalid? Return 400 Bad Request
    ▼
Authenticate user
    ├─ No token? Return 401 Unauthorized
    ├─ Invalid token? Return 401 Unauthorized
    ▼
Check admin status
    ├─ Not admin? Return 403 Forbidden
    ▼
Call Gemini API
    ├─ No key? Return 500
    ├─ API error? Return 500
    ├─ Timeout? Return 500
    ├─ Invalid response? Return 500
    ▼
Parse response
    ├─ Invalid JSON? Return 500
    ├─ Missing fields? Return 500
    ▼
Save to database
    ├─ DB error? Return 500
    ▼
Return 201 Created with question
```

## API Endpoints Map

```
PUBLIC ROUTES (/api/v1)
├── GET /health
├── GET /auth/methods
├── POST /auth/signup
├── POST /auth/signin
├── POST /auth/forgot-password
├── GET /auth/google
└── GET /auth/google/callback

PROTECTED ROUTES (/api/v1) - Requires Auth
├── GET /verify
├── GET /profile
├── PUT /users/profile
├── GET /puzzles
├── GET /puzzles/:id
├── GET /questions ................. NEW
├── GET /questions/:id ............. NEW
└── GET /questions/category/:category NEW

ADMIN ROUTES (/api/v1/admin) - Requires Auth + Admin
├── POST /users/make-admin
├── GET /users
├── POST /puzzles
└── POST /questions/generate ........ NEW
```

## Configuration Initialization Order

```
1. Load .env file
   config.LoadEnv()
   ▼
2. Initialize Gemini API
   config.InitializeGemini()
   │
   ├─ Read GEMINI_API_KEY
   ├─ Create genai.Client
   ├─ Store in config.GeminiClient
   ▼
3. Initialize Database
   database.InitDB()
   │
   ├─ Read database credentials
   ├─ Connect to PostgreSQL
   ├─ Run AutoMigrate
   │  ├─ Create/update users table
   │  ├─ Create/update puzzles table
   │  └─ Create/update questions table .... NEW
   ▼
4. Create Gin Router
   gin.Default()
   ▼
5. Initialize Firebase
   config.InitializeFirebase()
   ▼
6. Initialize Routes
   routes.InitializeRoutes()
   │
   ├─ Create controllers
   ├─ Create middleware
   ├─ Define all endpoints
   ▼
7. Start Server
   router.Run(":8080")
   ▼
8. Ready to receive requests
```

## Data Transformation Pipeline

```
GEMINI RESPONSE (JSON)
{
  "id": "xyz",
  "title": "...",
  "question": "...",
  "solution": {
    "answer": "...",
    "explanation": "..."
  },
  "hints": [...],
  "difficulty": "...",
  ...
}
    ▼
PARSING (parseGeneratedQuestion)
    │
    ├─ Extract text fields
    ├─ Convert arrays
    ├─ Add UUID if needed
    ├─ Validate required fields
    ▼
GO STRUCT (Question Model)
{
  ID: 1,
  QuestionID: "uuid",
  Title: "...",
  Question: "...",
  Answer: "...",
  Explanation: "...",
  Hints: {...},
  Difficulty: "...",
  CreatedBy: 1,
  ...
}
    ▼
DATABASE INSERT (GORM)
    ▼
POSTGRESQL ROW
questions table
    ▼
JSON RESPONSE
    ▼
FRONTEND
```

## Deployment Architecture

```
┌──────────────────────────────────────┐
│         Environment (.env)            │
│  GEMINI_API_KEY                      │
│  POSTGRES_*                          │
│  FIREBASE_*                          │
└──────────────────────────────────────┘
            ▼
┌──────────────────────────────────────┐
│  Go Application (thinkbattleground)  │
│  ├─ main.go                          │
│  ├─ config/                          │
│  ├─ controllers/                     │
│  ├─ models/                          │
│  ├─ routes/                          │
│  └─ database/                        │
└──────────┬──────────────────────────┘
           │
    ┌──────┴──────┐
    │             │
    ▼             ▼
┌─────────┐  ┌──────────────┐
│ Gemini  │  │ PostgreSQL   │
│  API    │  │  Database    │
└─────────┘  └──────────────┘
```

---

**Version:** 1.0.0
**Last Updated:** December 7, 2025
