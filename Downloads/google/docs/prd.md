# FinGuard AI — Product Requirements Document (PRD)

> Mirror of the source PRD. Source-of-truth lives in the originating brief; this file is for repo reviewers.

## 1. Executive Summary

**Product Name:** FinGuard AI
**Tagline:** Real-time fraud detection, risk intelligence, and personalized financial guidance powered by Google Cloud AI.

FinGuard AI is an AI-powered platform for banks, NBFCs, insurers, and fintechs that unifies fraud detection, risk assessment, document intelligence, and personalized financial assistance into a single intelligent layer. It combines real-time streaming analytics (Pub/Sub + BigQuery), predictive ML (Vertex AI), generative reasoning (Gemini), and document understanding (Document AI) to reduce fraud losses, accelerate compliance, and improve customer trust.

## 2. Problem Statement

| Challenge | Impact |
|---|---|
| Siloed fraud detection (rules-only) | High false positives (~90 %), missed novel fraud |
| Manual document processing (KYC, claims, loans) | Slow onboarding, human error, cost |
| Reactive risk models | Delayed response to emerging threats |
| Generic customer experiences | Low engagement, churn |
| Compliance overhead (AML, KYC, GDPR) | Regulatory penalties, audit fatigue |

## 3. Goals & Objectives

### Business Goals
- Reduce fraud losses by **40 %** within 12 months
- Cut false positives by **60 %** via ML-based scoring
- Reduce document processing time by **75 %** (KYC/claims/loan docs)
- Improve customer NPS by **+20 points**
- Automate **80 %** of compliance reporting workflows

### Product Goals
- Real-time transaction anomaly detection (< 500 ms latency)
- Unified risk score across customer, transaction, and merchant
- Conversational AI assistant for customers and analysts
- Explainable AI for regulatory audits

## 4. Target Users / Personas

| Persona | Needs |
|---|---|
| **Fraud Analyst** | Real-time alerts, case management, explainability |
| **Risk Officer** | Portfolio risk dashboards, predictive scores |
| **Compliance Officer** | AML/KYC automation, audit trails |
| **Customer** | 24/7 assistant, personalized advice, dispute tracking |
| **Ops / Back Office** | Document automation (invoices, claims, KYC) |
| **Data Scientist** | Model training, monitoring, drift detection |

## 5. Key Features

### 5.1 Real-Time Fraud & Anomaly Detection
- Streaming ingestion of transactions via Pub/Sub
- Feature enrichment in BigQuery (velocity, geo, device, merchant risk)
- Vertex AI anomaly detection model (autoencoder + gradient boosted trees)
- Sub-500 ms scoring via Cloud Run
- Auto-alerting, case creation, analyst review queue

### 5.2 Document Intelligence
- Document AI parses KYC, invoices, claims, loan applications
- Entity extraction: names, IDs, amounts, dates, signatures
- Auto-verification against internal + external sources
- Confidence scoring with human-in-the-loop fallback

### 5.3 Conversational Financial Assistant
- Gemini on Vertex AI for customers + analysts
- Natural language queries: "Why was my card declined?"
- Personalized recommendations (savings, insurance, investments)
- Multi-channel: web, mobile, chat, voice

### 5.4 Risk Assessment & Predictive Scoring
- Unified Customer Risk Score (0–1000)
- Portfolio-level risk dashboards in Looker
- Stress testing and scenario simulation

### 5.5 Compliance Automation
- AML / KYC / GDPR workflows
- Auto-generated SAR (Suspicious Activity Reports)
- Policy retrieval via Vertex AI Search
- Immutable audit trails

### 5.6 Knowledge & Policy Retrieval
- Vertex AI Search / Agent Search over internal policies
- Context-grounded answers for analysts and customers
- Citation-backed responses for compliance

## 6. Success Metrics (KPIs)

| Metric | Target |
|---|---|
| Fraud detection recall | > 95 % |
| False positive rate | < 5 % |
| P95 scoring latency | < 500 ms |
| Doc processing time | < 30 s / doc |
| Customer NPS | +20 |
| Analyst productivity | +40 % |

## 7. Scope

**In Scope (v1):**
- Retail banking transactions
- KYC + loan document processing
- Customer + analyst chat assistant
- AML compliance workflow

**Out of Scope (v1):**
- Investment trading
- Crypto custody
- Cross-border payments rails

## 8. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Model drift | Continuous monitoring + auto-retraining |
| False positives | Human-in-the-loop + feedback loops |
| Data privacy | CMEK, VPC-SC, data residency |
| Regulatory | Explainable AI + audit logs |
| Latency | Edge caching + regional deployment |

## 9. Release Plan

| Phase | Duration | Deliverable |
|---|---|---|
| Phase 1 | 0–8 weeks | Data pipeline + fraud MVP |
| Phase 2 | 8–16 weeks | Document AI + chat assistant |
| Phase 3 | 16–24 weeks | Risk dashboards + compliance |
| Phase 4 | 24–32 weeks | Scale + multi-region |
