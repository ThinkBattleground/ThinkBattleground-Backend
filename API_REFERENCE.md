# Question Generation API - Reference Card

## Quick Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
All endpoints require Firebase token in header:
```
Authorization: Bearer <firebase_token>
```

---

## Endpoints

### 1. Generate Question (Admin Only)

**URL:** `/admin/questions/generate`
**Method:** `POST`
**Auth:** Required (Admin)

**Request:**
```json
{
  "category": "algebra",
  "difficulty": "intermediate"
}
```

**Difficulty Options:**
- `beginner` - Easy, ~5 min, 10-30 points
- `intermediate` - Medium, ~10-15 min, 50-100 points
- `advanced` - Hard, ~20-30 min, 100-200 points
- `expert` - Very hard, 30+ min, 200+ points

**Response (201):**
```json
{
  "message": "Question created successfully",
  "question": {
    "id": 1,
    "question_id": "uuid-string",
    "title": "Question Title",
    "question": "Full question text",
    "answer": "Answer",
    "explanation": "How to solve it",
    "hints": ["Hint 1", "Hint 2"],
    "difficulty": "intermediate",
    "expectedTime": 10,
    "points": 50,
    "category": "algebra",
    "subcategory": "linear_equations",
    "tags": ["tag1", "tag2"],
    "requirements": ["req1"],
    "imageUrl": ""
  }
}
```

**Errors:**
- `400` - Invalid category/difficulty
- `401` - Not authenticated
- `403` - Not admin
- `500` - API error

---

### 2. List All Questions

**URL:** `/questions`
**Method:** `GET`
**Auth:** Required

**Response (200):**
```json
{
  "count": 5,
  "questions": [
    { /* question objects */ }
  ]
}
```

---

### 3. Get Single Question

**URL:** `/questions/:id`
**Method:** `GET`
**Auth:** Required
**Param:** `id` = question_id (UUID)

**Response (200):**
```json
{
  /* full question object */
}
```

**Errors:**
- `404` - Question not found

---

### 4. Get Questions by Category

**URL:** `/questions/category/:category`
**Method:** `GET`
**Auth:** Required
**Param:** `category` = math category

**Response (200):**
```json
{
  "category": "algebra",
  "count": 3,
  "questions": [
    { /* question objects */ }
  ]
}
```

---

## Category List

```
algebra              - Linear, abstract, equations
calculus             - Differential, integral
geometry             - Euclidean, analytical
trigonometry         - Trig functions, identities
statistics           - Data, distributions
probability          - Events, random variables
arithmetic           - Basic math, numbers
discrete_math        - Logic, combinatorics
linear_algebra       - Matrices, vectors
number_theory        - Prime, divisibility
```

---

## cURL Examples

### Generate Question
```bash
curl -X POST http://localhost:8080/api/v1/admin/questions/generate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"category":"algebra","difficulty":"beginner"}'
```

### List Questions
```bash
curl -X GET http://localhost:8080/api/v1/questions \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get by ID
```bash
curl -X GET http://localhost:8080/api/v1/questions/123e4567-e89b-12d3-a456-426614174000 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get by Category
```bash
curl -X GET http://localhost:8080/api/v1/questions/category/algebra \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success - GET request |
| 201 | Created - POST request |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - No/invalid token |
| 403 | Forbidden - Not admin |
| 404 | Not Found - Resource doesn't exist |
| 500 | Server Error - API/DB issue |

---

## Question Object Structure

```typescript
{
  id: number,                    // Database ID
  question_id: string,           // UUID (unique)
  title: string,                 // Question title
  question: string,              // Question text
  answer: string,                // Final answer
  explanation: string,           // Solution steps
  hints: string[],               // 2-3 hints
  difficulty: string,            // beginner|intermediate|advanced|expert
  expectedTime: number,          // Minutes
  points: number,                // Points for solving
  category: string,              // Main category
  subcategory: string,           // Specific topic
  tags: string[],                // Concepts used
  requirements: string[],        // Prerequisites
  imageUrl: string,              // Diagram URL (optional)
  created_by: number,            // Creator user ID
  created_at: string,            // ISO timestamp
  updated_at: string,            // ISO timestamp
}
```

---

## Error Response Format

All errors return:
```json
{
  "error": "Error description"
}
```

Common errors:
```json
// Invalid category
{ "error": "Invalid difficulty level. Must be: beginner, intermediate, advanced, or expert" }

// Not authenticated
{ "error": "User not found" }

// Not admin
{ "error": "Unauthorized: Admin privileges required" }

// Not found
{ "error": "Question not found" }

// Server error
{ "error": "Failed to generate question: [reason]" }
```

---

## Rate Limiting

Limits based on Gemini API:
- Google's free tier: Limited requests/minute
- Paid tier: Higher limits
- Check quota at: https://console.cloud.google.com/apis/api/generativelanguage.googleapis.com/quotas

---

## Setup Checklist

- [ ] Firebase token obtained
- [ ] Authorization header formatted correctly
- [ ] Base URL correct
- [ ] Gemini API key configured
- [ ] Server running on port 8080

---

## Testing Workflow

1. **Generate Questions** (admin)
   ```bash
   # Create a few questions
   curl -X POST .../admin/questions/generate ...
   ```

2. **List Questions** (user)
   ```bash
   # Verify they were saved
   curl -X GET .../questions ...
   ```

3. **Retrieve by Category** (user)
   ```bash
   # Filter results
   curl -X GET .../questions/category/algebra ...
   ```

4. **Get Single Question** (user)
   ```bash
   # View full details
   curl -X GET .../questions/uuid ...
   ```

---

## Field Validation

| Field | Required | Type | Max Length | Values |
|-------|----------|------|-----------|--------|
| category | Yes | string | - | Any math domain |
| difficulty | Yes | string | - | beginner, intermediate, advanced, expert |
| title | No | string | 255 | - |
| question | No | string | - | - |
| answer | No | string | - | - |
| hints | No | array | - | 2-3 items |
| points | No | number | - | Positive |
| expectedTime | No | number | - | Minutes |

---

## Performance Tips

1. **Use pagination for large result sets** (future enhancement)
2. **Cache questions in frontend** (reduce API calls)
3. **Generate in batch** (reduce API quota usage)
4. **Filter by category** (reduce data transfer)
5. **Monitor API quotas** (avoid surprises)

---

## Integration Example

```javascript
// Frontend integration
const token = localStorage.getItem('firebaseToken');

// Generate question
const response = await fetch(
  'http://localhost:8080/api/v1/admin/questions/generate',
  {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      category: 'algebra',
      difficulty: 'intermediate'
    })
  }
);

const data = await response.json();
console.log(data.question);
```

---

## Support

For issues or questions:
1. Check `QUESTION_GENERATION.md` for details
2. See `TEST_QUESTIONS.md` for examples
3. Review `IMPLEMENTATION_SUMMARY.md` for architecture
4. Check Gemini API docs at https://ai.google.dev

---

**Last Updated:** December 7, 2025
**API Version:** 1.0.0
