# 1. Introduction

## 1.1 Overview

OpenBPL (Open Brand Protection Library) is an open-source framework for monitoring, detecting, and acting against brand infringements across the web. This document outlines the functional and technical requirements to implement OpenBPL’s features effectively.

## 1.2 Objectives

- Implement an automated brand monitoring and takedown system using OpenBPL.
- Enable scalable integration with external APIs and data sources.
- Provide real-time insights into brand infringements and proactive response mechanisms.
- Leverage Large Language Models (LLMs) for threat classification and automated legal document generation to reduce manual effort.
- Ensure a user-friendly interface for analysts to monitor and manage threats.
- Support self-hosting capabilities for easy deployment with minimal dependencies.

## 1.3 Key Stakeholders

- CTO / Engineering Team: Responsible for initial deployment and maintenance of OpenBPL instance.
- Legal / Compliance Team: Manages takedown workflows and legal documents.
- Marketing & Product Teams: Provides brand assets and intel on new launches and campaigns.
- Brand Protection Analysts: Monitors and responds to threats.

# 2. Product Scope

## 2.1 Core Functionalities

### 2.1.1 Threat Detection:

- Monitor social media, domains, certificates, marketplaces, and other digital spaces.
- Use AI-driven heuristics and pattern recognition to detect infringements.
- Enable LLM-assisted classification of threats.

### 2.1.2 Threat Enrichment & Analysis:

- Aggregate data from multiple scanning sources.
- Automatically generate threat reports with metadata, screenshots, and historical logs.
- Implement risk scoring using LLM-driven analysis.

### 3. Automated Takedown Workflows:

- Identify hosting providers, domain registrars, and platform contacts.
- Automate legal notices and DMCA takedowns.
- Track and log the status of takedown requests.

### 4. Dashboard & Reporting:

- Interactive UI for monitoring threats.
- Customizable alerts and real-time notifications.
- Exportable reports for compliance and audits.

### 5. API & Integrations:

- OpenAPI support for integrating external tools.
- Webhook-based notifications for third-party services.
- API key management for secure access to data and functionality.
- LLM integration for automated classification and decision-making.
- Support for multiple LLM providers (OpenAI, Llama, Claude, Gemini, etc.).

### 6. User Management & Security:

- Role-based access control (RBAC) for different user roles.
- Secure authentication mechanisms (OAuth, API keys).
- Audit logs for tracking user actions and system events.

# 3. Technical Requirements

## 3.1 Infrastructure

- Deployment: Simple Docker image (with docker-compose.yml file for services).
- Database: In-memory for development, PostgreSQL for production.
- Storage: Local filesystem for development, Object storage (e.g., AWS S3, MinIO) for production.
- Monitoring: Prometheus metrics and OLTP for observability.
- Email: SMTP server for sending notifications.
- Webhooks: Support for third-party integrations (Slack, Discord, Telegram).
- API: OpenAPI specification for all endpoints.
- Browser Automation: Use Playwright or Puppeteer for web scraping and screenshot capture.
- Background Jobs: [River Queue](https://riverqueue.com/) for processing tasks asynchronously.

## 3.2 Security & Compliance

- Role-based access control (RBAC) for different stakeholders.
- Secure API authentication (OAuth, API keys).
- Audit logs for all events.

## 3.3 Performance & Scalability

- Support real-time scanning with event-driven architecture.
- Scalable queue-based processing.
- Rate-limiting and load balancing to handle API limits.
- Ability to scale horizontally by deploying separate worker nodes.

# 4. LLM Integration Requirements

## 4.1 Supported LLMs

- OpenAI (GPT-4), Llama, Claude, Gemini, or self-hosted models.
- Fine-tuned models for legal text generation and threat classification.

## 4.2 LLM Implementation

- Classification Tasks: Assign risk scores to threats.
- Automated Legal Notices: Generate and refine takedown requests.
- Conversational Interface: AI assistant for querying threats and running scans.

## 4.3 Configuration & Customization

- API-based LLM selection.
- Custom prompt engineering for use-case-specific responses.
- Toggle between different LLM providers.
- Ability to bring your own keys (BYOK)

# 5. User Workflows

## 5.1 Brand Management

1. Marketing team provides assets (logos, trademarks, keywords).
2. OpenBPL enriches brand data with automated scraping and LLM analysis.
3. Analysts can add new brands and manage existing ones through the UI or API.

## 5.2 Threat Detection

1. Scanner identifies a potential infringement.
2. OpenBPL fetches metadata, screenshots, and domain info.
3. LLM assigns a risk score and classification.
4. Analysts review and escalate for takedown.

## 5.3 Automated Takedown

1. Identified infringement triggers automated action.
2. OpenBPL sends legal notices to the hosting provider.
3. System tracks response times and resolution.

## 5.4 Monitoring & Reporting

1. Analysts access the dashboard to view active threats.
2. Customizable alerts notify stakeholders of new threats.
3. Exportable reports for compliance and audits.
4. Email notifications for critical events.
5. Webhook notifications for third-party integrations (e.g., Slack, Discord, Telegram).

## 5.5 False Positive Management

1. Analysts review flagged threats.
2. LLM assists in classifying false positives.
3. System learns from analyst feedback to improve future classifications.
4. Takedown requests can be manually retracted if false positives are identified.
5. Analysts can provide feedback to the LLM to refine its classification capabilities.
6. System logs all false positive cases for auditing and training purposes.

# 6. Development Phases

## 6.1 Phase 1: Detection Vertical Slice and CLI interface

- Implement a vertical slice of the threat detection pipeline.
- CLI interface for running scans and managing threats.
- Scan new domains from certificate transparency logs.
- Process domains and extract keywords.
- Use browser automation to fetch HTML content and screenshots.
- Implement favicon similarity checks.

## 6.2 Phase 2: Monitoring Multiple Brands

- Extend the system to monitor multiple brands.
- YAML configuration for brand assets and keywords.
- Email and webhook notifications for new threats.

## 6.3 Phase 3: Detections as Code Engine

- Implement a code engine for defining detection rules.
- Support for custom detection rules using DSL (Domain-Specific Language).
- Integration with LLM for rule generation and generation of realistic test cases.
- Automated threat enrichment with screenshots and metadata (e.g., WHOIS, DNS records).
- DAG (Directed Acyclic Graph) for managing dependencies between detection rules and enrichment tasks.

## 6.4 Phase 4: Cost-based Analysis

- Implement cost-based analysis for detection rules.
- Use LLM to analyze the cost of running detection rules.
- Provide insights into the most effective rules based on cost and detection rate.
- Add detection results to the database for historical analysis.
- Schedule periodic scans for continuous monitoring based on LLM predictions of threat likelihood.
- Implement a feedback loop where analysts can mark detections as false positives or true positives, allowing the system to learn and improve over time.

## 6.5 Phase 5: Takedown Automation

- Implement automated takedown workflows.
- Integrate with LLM for generating legal notices and DMCA takedowns.
- Support for multiple LLM providers for legal document generation.
- Implement a user-friendly interface for analysts to manage takedown requests.
- Track the status of takedown requests and provide real-time updates.
- Implement a notification system for takedown request status changes (e.g., successful, failed, pending).
- Enable analysts to review and approve automated takedown requests before sending.
- Implement a refiling mechanism for takedown requests that were not successful.

# 7. Success Metrics

- Detection Rate: % of threats correctly identified.
- Takedown Efficiency: Time taken from detection to removal.
- False Positive Rate: Accuracy of LLM-based classification.
- User Adoption: Active engagement by analysts.
- System Performance: Response time for API calls and dashboard interactions.
- Scalability: Ability to handle increased load without performance degradation.
- Integration Success: Number of successful third-party integrations (e.g., webhooks, email notifications).
- Compliance: Adherence to legal and security standards.

# 8. Open Questions

1. How to ensure scalability for large-scale monitoring while keeping things easily self-hostable?
2. What customization options should be offered for enforcement workflows?
3. How to handle rate limits and API restrictions from third-party services?
4. What are the best practices for managing LLM keys and ensuring secure access?
5. How to balance between automated and manual processes in takedown workflows?
6. How to handle false positives effectively while minimizing manual intervention?
7. How to handle large volumes of data from multiple scanning sources without performance degradation?
