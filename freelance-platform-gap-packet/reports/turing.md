# Turing Live-Form Evidence Report

## Evidence boundary

- **Retrieved:** 2026-07-24 14:31-14:35 CDT (19:31-19:35 UTC).
- **Method:** Public Turing pages, unauthenticated routes, and current Turing-served production modules. No account, email verification, OAuth continuation, form submission, assessment, purchase, or verification bypass.
- **Rule:** Production modules establish current client validation but do not prove that every conditional field appears for every authenticated applicant.

## Source register

| ID | URL | Observed result |
|---|---|---|
| S1 | https://www.turing.com/jobs | Public talent landing page; Apply Now targets `/signup`. |
| S2 | https://developers.turing.com/signup | HTTP 200, public client-rendered signup. |
| S3 | https://developers.turing.com/welcome/ | HTTP 200 shell; client route guard redirects unauthenticated users to login. |
| S4 | https://developers.turing.com/_next/static/chunks/a099cd21902352d0.js | Current signup implementation. |
| S5 | https://developers.turing.com/_next/static/chunks/3bd1881a6780044e.js | Current resume-step implementation. |
| S6 | https://developers.turing.com/_next/static/chunks/7237b74585296e28.js | Current basic-information UI. |
| S7 | https://developers.turing.com/_next/static/chunks/c9221468693dd53f.js | Current profile validators. |
| S8 | https://developers.turing.com/_next/static/chunks/571361aed8053ae2.js | Current education implementation. |
| S9 | https://developers.turing.com/_next/static/chunks/014fed445fa2d6e5.js | Current role, skills, and languages implementation. |
| S10 | https://developers.turing.com/_next/static/chunks/4ac082a1a9de8e81.js | Current work-history/profile-step implementation. |
| S11 | https://developers.turing.com/_next/static/chunks/11bbbf73984d1cfd.js | Current professional-links implementation. |
| S12 | https://www.turing.com/policy | Current privacy policy, last updated 2025-06-27. |

## Directly observed signup

The public signup offers **Continue with Google** or a single required **Email** field and **Sign up** action. Blank input reports `Required`; invalid syntax reports `Invalid email address`. No password field is displayed. Continuing represents agreement to Turing's Terms of Service and Privacy Policy. The current implementation sends the email and an empty password value, then has an activation-message state for the verification email. [S2][S4]

The welcome, profile, jobs, and verification URLs can return HTTP 200 application shells while still being client-protected. HTTP 200 is not evidence of public form access. [S3]

## Authenticated profile requirements exposed by current modules

These controls are not interactively accessible without authentication, but Turing's current production modules expose their rules.

### Resume

- Upload is **recommended but skippable**.
- PDF, `.doc`, and `.docx`; maximum 4 MiB.
- Parsing can autofill skills and education. Replacing a resume clears completed fields. [S5]

### Basic information

Required: first name (2-50 characters), last name (2-50), country of residence, phone number, and either an individual LinkedIn profile URL or **I don't have a LinkedIn profile**. Phone accepts at least six characters from digits, spaces, hyphens, parentheses, and optional leading `+`. Optional/conditional controls include state/province, city, nationality, and non-negative expected annual earnings in local currency. [S6][S7]

### Education

At least one entry is required. School, degree, and field of study are required. Dates are conditional; end cannot precede start. CGPA and scale are optional as a pair and CGPA cannot exceed scale. A **Currently studying** state is available. [S8][S7]

### Role, skills, and languages

Required: full-time work experience as an integer from 0-50, preferred role, at least one skill, 1-50 years of professional experience for each selected skill, and at least one language. Up to five languages may be added; speaking/reading/writing proficiency is optional. Turing says selected skills drive technical assessment focus. [S9][S7]

### Work history and projects

Each work entry requires position, company, start year, and end year/current status. Added projects require name, start year, end year/current status, and project details. Project URL is optional but must be HTTP/HTTPS if supplied; skills and attachments are supported. [S10][S7]

### Professional links and remaining steps

Optional link types include GitHub, Google Scholar, ResearchGate, Dribbble, and Other. Added links require type and valid HTTP/HTTPS URL; Other also requires a title. GitHub is not universal. The profile controller also names Publications, Profile summary, and final **Submit Profile**, but exact conditional requiredness is inaccessible without an authenticated profile state. [S10][S11]

## Vetting, documents, and inaccessible facts

Current public guidance supports profile creation, role recommendations/applications, then skill assessments and live interviews matched to the applicant's experience. It does not establish one universal sequence, test count, duration, passing score, retake interval, proctoring setup, or one-hour stack interview. [S1]

Turing's privacy policy says recorded assessments may collect qualifications/experience and live coding interviews may record name, image, likeness, identification information, voice, screen sharing, and problem-solving activity. It says Turing **may** collect government identification and background-check information. This does not prove either is universal. [S12]

Exact role questions, publication/summary requiredness, matching time, compensation, weekly-hours requirement, time-zone overlap, client interview count, and post-submission status remain inaccessible or role-specific.

## User-only actions and exact gaps

Solomon must personally:

- Choose email or Google, submit signup, and complete account/email verification.
- Upload or deliberately skip a truthful current resume, then verify any parsed content.
- Supply legal/contact, education, role, skill, language, work-history, project, and summary data shown in the authenticated profile.
- Decide whether to provide LinkedIn, portfolio links, expected earnings, and optional fields.
- Inspect each role's actual questions before applying.
- Complete any assigned assessment, recorded/live interview, permissions, ID/background check, contract, payment, or onboarding flow.

**Known material gap:** no resume/CV exists in the reviewed Hermes artifact tree. Existing profile copy does not provide exact education dates/degree/field, legal contact details, LinkedIn decision/URL, or exact years per selected skill.

## Reconciliation of prior assumptions

| Prior claim | Current determination |
|---|---|
| Signup requires email/password | **Incorrect.** Initial signup displays email only or Google. |
| Resume is required | **Incorrect.** Current step is skippable. [S5] |
| GitHub/portfolio is required | **Incorrect.** Professional links are optional types. [S11] |
| Universal Work Experience Survey, 57 questions, coding tests, live interview, and one-hour stack interview | **Unsupported as a universal current flow.** Current wording is role-matched. [S1] |
| Universal 40-hour commitment and US overlap | **Unverified.** Role-specific requirements may vary. |
| Vetting takes 3-6 hours and matching 2-4 weeks | **Unverified by current primary requirements.** |
| Universal ID/background check | **Not established.** Privacy policy says Turing may collect these. [S12] |
| $100-$200/hour client rates or a known developer split | **Unsupported by current primary application evidence.** |

## Current readiness

**Observed status: Not submitted.** Public and module-exposed requirements are reconciled. The authenticated profile, role selection, submission, and any assigned vetting remain user-only.
