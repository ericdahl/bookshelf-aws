# Book Tracker: AWS Serverless App

## ✨ Overview

This is a fully serverless application that allows users to track books they are currently reading, have read, or want to read. The application uses modern AWS serverless technologies including API Gateway, Lambda, DynamoDB, Cognito, S3 (with signed URLs), and optionally AWS Bedrock for AI-powered recommendations.

The goal is to provide a complete reference implementation for developers looking to learn and deploy real-world serverless applications on AWS.

---

## 📆 Features

### Core

* User sign-up and authentication via Amazon Cognito
* Track books by status:

  * Want to Read
  * Currently Reading
  * Finished Reading
* Store book metadata: title, author, status, rating, review, tags
* Allow users to update or delete book entries

### Reports (S3 Signed URLs)

* Generate a downloadable reading summary in CSV format
* File stored in S3 with private permissions
* Return signed S3 URL to client for secure download

### Book Search

* Search for books using Google Books API (title, author, etc.)
* Auto-fill book metadata when adding a new book
* Fallback to manual entry if desired

### AI Integration (AWS Bedrock)

* Use Bedrock (e.g. Claude or Titan) to generate personalized book recommendations
* Prompt based on user's reading history and preferences
* Return results with titles, authors, and genres

### Optional Enhancements

* CSV import of book list with signed upload URL
* Audio review uploads to S3 (voice memos)
* Periodic email report with reading history (via SES)

---

## 🛁 Architecture

### Overview

```
Frontend (React/S3/CloudFront)
    |
    |---> API Gateway (HTTP API + JWT Authorizer)
             |
             |---> Lambda Functions
             |        |- Book CRUD
             |        |- Book Search Proxy (Google Books)
             |        |- Recommendation Engine (Bedrock)
             |        |- Report Generator (S3 Signed URL)
             |
             |---> DynamoDB (Books Table)
             |---> S3 (Reports Bucket)
             |---> Cognito (User Auth)
```

---

## 📊 Data Model

### DynamoDB: `Books` Table

Partition Key: `PK = USER#<user_id>`
Sort Key: `SK = BOOK#<book_id>`

Attributes:

```json
{
  "title": "The Hobbit",
  "author": "J.R.R. Tolkien",
  "status": "reading",
  "rating": 5,
  "review": "Epic fantasy classic",
  "tags": ["fantasy", "classic"],
  "started_at": "2025-06-01",
  "finished_at": null
}
```

### GSI for Status-Based Queries

* GSI1 Partition Key: `status`
* GSI1 Sort Key: `PK`

---

## 🔧 API Endpoints

### Book Management

```
GET    /books              --> List all books
GET    /books?status=...  --> Filter by status
POST   /books              --> Create new book
PUT    /books/{id}         --> Update book
DELETE /books/{id}         --> Delete book
```

### Reports

```
POST   /report             --> Generate CSV report, return signed S3 URL
```

### Search

```
GET    /search?q=the+hobbit   --> Proxy to Google Books API, return suggestions
```

### Recommendations

```
GET    /recommendations       --> Bedrock prompt based on reading history
```

---

## 🛡 IAM & Security

* All API endpoints protected via Cognito JWT authorizer
* Users only access their own books via IAM policies + DynamoDB conditions
* S3 bucket private, files accessed via short-lived signed URLs
* Lambda has scoped-down permissions (DynamoDB, S3, Bedrock)

---

## 🚀 Deployment

Recommended: [AWS SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam.html) or [AWS CDK](https://docs.aws.amazon.com/cdk/)

* `sam build && sam deploy` for end-to-end setup
* Use environment variables for API keys (e.g. Google Books)

---

## 📁 Directories & Files

```
/infra               --> CDK or SAM templates
/backend             --> Lambda handlers (Python or Node.js)
/frontend            --> React app (optional)
/docs                --> Diagrams, design notes
/tests               --> Unit and integration tests
```

---

## 📖 External APIs

### Google Books API

* Endpoint: `https://www.googleapis.com/books/v1/volumes?q=...`
* No auth needed for basic usage
* Rate-limiting may apply; use backend proxy if needed

### AWS Bedrock

* Use Claude or Titan via `bedrock:InvokeModel`
* Construct a prompt like:

  > A user has read: Dune, The Martian. Recommend 5 similar books with genre.

---

## ✅ Future Enhancements

* PDF report generation (HTML-to-PDF via Puppeteer)
* SES email report delivery
* Book club / shared lists between users
* Mobile frontend (React Native or Flutter)
* Tech Debt
  * avoid hard-coding endpoints
  * cloudfront
  * common JWT library/logic
  * cache-control headers on S3/API


---

## 🙏 Contributions

This is intended as a hands-on learning project. Contributions welcome via pull request or issue.

---

## ⚡ Credits

Created as a learning tool to explore AWS Serverless, Bedrock, and best practices in full-stack development.

---

## ✨ License

MIT License (or choose your own)
