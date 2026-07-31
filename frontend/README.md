# Address Quality Frontend

This directory contains the frontend for the Address Quality project, including the landing page, documentation site, and interactive API playground.

Unlike the backend, which contains the core address validation engine, this application focuses on providing a clean developer experience for exploring and integrating the API.

---

## Development Philosophy

The frontend is intentionally developed using an **AI-assisted ("vibe coding") workflow**.

Address Quality is fundamentally a backend and infrastructure project. Instead of investing significant time building frontend components from scratch, the UI is rapidly prototyped using modern AI coding assistants and then manually reviewed, refined, and integrated into the project.

Every generated change is reviewed before being committed. AI is used to accelerate implementation—not to replace engineering judgment.

This allows development effort to remain focused on the parts of the project that provide the most value:

- Indonesian address parsing
- Administrative hierarchy validation
- Candidate resolution
- Confidence scoring
- Explainable validation evidence

If you're a frontend engineer and notice opportunities to improve the codebase, contributions are always welcome.

---

## Tech Stack

- React
- TypeScript
- Vite
- Tailwind CSS
- Kumo UI

---

## Getting Started

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run dev
```

Build for production:

```bash
npm run build
```

Preview the production build:

```bash
npm run preview
```

---

## Project Goals

The frontend aims to provide:

- A modern landing page
- Comprehensive API documentation
- Interactive API playground
- Responsive developer experience
- Clean and maintainable UI components

It is intentionally lightweight so the majority of development effort can remain focused on the Address Quality API and its validation engine.