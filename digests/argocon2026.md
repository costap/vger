# Playlist Digest — PLj6h78yzYM2PM2c_Kv-vx8KB9sI2y7NM2

*31 talks analysed by V'Ger*

## Overarching Theme

This playlist deeply explores the evolving Argo ecosystem, showcasing its foundational role in enabling advanced GitOps, progressive delivery, and scalable workflow automation on Kubernetes. It highlights significant advancements in developer experience, security, and performance, increasingly integrating AI/ML pipelines and tackling complex multi-cluster, hybrid cloud, and FinOps challenges while fostering a robust open-source community.

## Watch These First

1. **[ArgoCon | Argo Project Velocity Update - Dan Garfield](https://www.youtube.com/watch?v=8nb6OcMmgmQ)**  
   This talk serves as an excellent high-level introduction to the entire Argo project, its core components (Workflows, CD, Rollouts, Events), and its significant community growth and relevance in areas like AI/data pipelines.

2. **[ArgoCon | Sponsored Keynote: From Kubernetes to Anything: The Evolution of Promotion... Jesse Suen](https://www.youtube.com/watch?v=60xlbEzJiVQ)**  
   This keynote introduces Kargo, a significant new project that extends Argo CD to provide comprehensive promotion orchestration beyond simple deployments, offering a vision for the future of GitOps pipelines and enforcement boundaries.

3. **[Anatomy of a GitOps Engine: Understanding Your Sync - Alexandre Gaudreault, Intuit](https://www.youtube.com/watch?v=0yy4R6KfZSA)**  
   This talk provides a crucial deep dive into the internal workings of Argo CD's GitOps engine, explaining synchronization, diffing, and advanced features like sync hooks and waves, which are fundamental for troubleshooting and optimizing deployments.

4. **[GitOps and Secrets: State of the Union - Kostis Kapelonis, Octopus Deploy](https://www.youtube.com/watch?v=6d4dmR7pcPE)**  
   Addressing a critical and often mishandled aspect of GitOps, this talk offers clear guidance on securely managing secrets using external operators, a non-negotiable best practice for any production GitOps environment.

5. **[Don't Do What Donny Don't Does: The 10 Dos and 500 Don’ts of Workflows... Tim Collins & Becky Pauley](https://www.youtube.com/watch?v=tzCGRezWJzQ)**  
   This talk provides essential best practices and highlights common pitfalls for effectively using Argo Workflows at scale, which is vital for ensuring efficient, cost-effective, and reliable automation of complex pipelines.

## Recommended Learning Path

1. Kubernetes
2. GitOps
3. Argo CD
4. Argo Rollouts
5. Argo Workflows
6. Argo Events
7. Argo CD ApplicationSet
8. External Secret Operator (ESO)
9. Prometheus
10. Kargo

## Key Insights

The Argo ecosystem (Argo CD, Workflows, Rollouts, Events) is a mature, actively developed, and comprehensive suite for cloud-native automation, covering declarative deployments (GitOps), advanced progressive delivery, complex workflow orchestration (especially for data/AI), and event-driven systems, all underpinned by a strong community. GitOps is evolving beyond mere deployment to encompass full promotion orchestration and intelligent operations, with new projects like Kargo managing complete software promotion pipelines and emerging AI integrations enabling intelligent drift detection, self-healing progressive delivery, and auto-generated code fixes. Scalability, security, and cost optimization are paramount for successful GitOps adoption, emphasizing the need for optimizing Argo CD for large-scale environments (e.g., with ApplicationSets, efficient polling), securely managing secrets with external operators, and leveraging Argo Workflows for FinOps automation. Effective implementation requires a deep understanding of Kubernetes fundamentals and adherence to best practices, including proper resource allocation, observing cluster health, designing declarative configurations, and avoiding anti-patterns for reliability and maintainability. Finally, enhancing developer experience and fostering a GitOps culture are as important as technical implementation, with tools like Argo CD Diff Preview accelerating feedback loops and panels highlighting the organizational and cultural shifts needed to empower teams and build trust in automated delivery.

## Technology Radar

| Technology | Talks |
|---|---|
| Argo CD | 21 |
| Kubernetes | 16 |
| Argo Rollouts | 11 |
| Argo Workflows | 10 |
| Argo Events | 6 |
| Prometheus | 5 |
| GitOps | 3 |
| Helm | 3 |
| Kubernetes (K8s) | 3 |
| Terraform | 3 |
| Git | 2 |
| GitHub Actions | 2 |
| Istio | 2 |
| Kargo | 2 |
| Kustomize | 2 |
| LLMs (Large Language Models) | 2 |
| Open Policy Agent (OPA) | 2 |
| AI for Ops / AI Agents | 1 |
| Agent-to-Agent (A2A) | 1 |
| AnalysisTemplates | 1 |

## All Talks

### 1. ArgoCon | Argo Project Velocity Update - Dan Garfield

[https://www.youtube.com/watch?v=8nb6OcMmgmQ](https://www.youtube.com/watch?v=8nb6OcMmgmQ) · 11m 6s

This talk provides an update on the Argo Project's velocity, highlighting its robust community, significant growth, and recent achievements. The speaker, Dan Garfield, introduces the core Argo projects – Workflows, CD, Rollouts, and Events – emphasizing their increasing relevance, particularly Argo Workflows in AI/data pipelines. Key statistics reveal a thriving open-source ecosystem with over 2,000 contributors and an improved average merge time of 7 days, demonstrating efficiency and active development within the CNCF framework.

**Technologies:** Argo Workflows, Argo CD, Argo Rollouts, Argo Events, GitOps Controller (Argo Project Labs)

### 2. ArgoCon | Argo CD Update - Alexander Matyushentsev

[https://www.youtube.com/watch?v=iLSblVlU9cQ](https://www.youtube.com/watch?v=iLSblVlU9cQ) · 8m 20s

This talk provides a comprehensive update on Argo CD's development over the past year, highlighting four major releases (v3.0 to v3.3). Key advancements include enhanced security and performance defaults in v3.0, new extensibility features like OCI registry support and CLI plugins in v3.1, significant performance and observability improvements such as the Source Hydrator in v3.2, and critical synchronization optimizations for large-scale GitOps environments with server-side apply and shallow cloning in v3.3. The speaker emphasizes the robust community contributions that drive these innovations, focusing on both new features and essential project maintenance.

**Technologies:** Argo CD, OCI Registry Support (in Argo CD), Source Hydrator (in Argo CD), Parameterized UI Actions (in Argo CD), CLI Plugin Support (in Argo CD)

### 3. ArgoCon | Argo Rollouts Update - Alexander Gaudreault

[https://www.youtube.com/watch?v=vZ6hcrP_Wzc](https://www.youtube.com/watch?v=vZ6hcrP_Wzc) · 5m 20s

Argo Rollouts 1.9, now Generally Available, introduces the Progress Threshold feature, enhancing flexibility for canary deployments by allowing progress even if not all pods are immediately available. A major architectural shift towards a plugin-driven model has been implemented, enabling extensible integrations for traffic routing, metrics, and canary steps. This strategic move aims to foster community innovation and provide stable core functionality, while future plans include progressive canary deployments for stateful workloads like StatefulSets and DaemonSets.

**Technologies:** Argo Rollouts, RolloutPlugin controller, Kubernetes Gateway API

### 4. ArgoCon| Argo Workflows Update - Isitha Subasinghe

[https://www.youtube.com/watch?v=NmcaTwGcxxw](https://www.youtube.com/watch?v=NmcaTwGcxxw) · 4m 57s

This talk provides an update on Argo Workflows 4.0, highlighting significant enhancements like artifact plugins for custom storage backends, CEL-based CRD validation for earlier error detection, and operational improvements such as live parallelism tuning and context-aware structured logging. The release also includes the deprecation of singular mutex, semaphore, and schedule fields, with a convenient `argo convert` tool for migration, and a more robust and documented GoLang SDK for programmatic integration. These updates aim to enhance flexibility, improve user experience through faster feedback on errors, and streamline operations and development.

**Technologies:** Argo Workflows, Argo Workflows Artifact Plugins, Kubernetes Custom Resource Definitions (CRDs), CEL (Common Expression Language), Argo Convert CLI Tool

### 5. ArgoCon | Argo Events Update - Caelan Urquhart

[https://www.youtube.com/watch?v=uiWU3U7Aw1M](https://www.youtube.com/watch?v=uiWU3U7Aw1M) · 3m 15s

Caelan Urquhart from Pipekit presented an update on Argo Events, highlighting three recent releases that focus on improved connectivity, tuning capabilities, and operational reliability. Key enhancements include broader event intake with new GitLab emoji reaction triggers, enhanced operator control over Kafka partitioning, and sophisticated filtering logic with timezone awareness. These updates aim to make Argo Events more robust and user-friendly for large-scale, distributed systems.

**Technologies:** Argo Events, Argo Workflows, Kafka, GitLab, Azure Service Bus

### 6. ArgoCon l Sponsored Keynote: From Kubernetes to Anything: The Evolution of Promotion... Jesse Suen

[https://www.youtube.com/watch?v=60xlbEzJiVQ](https://www.youtube.com/watch?v=60xlbEzJiVQ) · 5m 59s

This talk introduces Kargo, a new project that extends Argo CD to provide comprehensive promotion orchestration beyond simple image tag updates. Kargo enables engineers to define sophisticated state-to-stage pipelines with custom verification gates, incorporating tools for supply chain security, policy validation, and infrastructure automation. By containerizing any tool into a 'Custom Promotion Step,' Kargo transforms promotion into an enforcement boundary, ensuring that only safe and compliant changes are pushed to Git and subsequently deployed by Argo CD. This approach solves how software 'moves' through environments, complementing Kubernetes' role in how software 'runs' and Argo CD's role in how software 'deploys.'

**Technologies:** Kargo, Argo CD, Kubernetes, Trivy, Open Policy Agent (OPA)

### 7. ArgoCon l Sponsored Keynote: Argo CD vs the World - Steve Fenton, Octopus Deploy

[https://www.youtube.com/watch?v=6G2Rr2rQ2RI](https://www.youtube.com/watch?v=6G2Rr2rQ2RI) · 5m 53s

Steve Fenton from Octopus Deploy presented an analysis of the 'State of GitOps Report,' comparing the performance of Argo CD users against non-Argo CD users across key DevOps metrics. The findings indicated that teams leveraging Argo CD consistently achieved higher GitOps scores, more frequent deployments, shorter lead times for changes, and faster recovery times from incidents. While change failure rates were comparable and generally low for both groups, Argo CD emerged as a clear winner in enhancing critical delivery pipeline health indicators.

**Technologies:** Argo CD, GitOps, DORA Metrics

### 8. Bring Your Own Artifact Driver To Workflows - Alan Clucas, Pipekit

[https://www.youtube.com/watch?v=kF6zuGZAgI0](https://www.youtube.com/watch?v=kF6zuGZAgI0) · 21m 53s

This talk introduces Artifact Driver Plugins in Argo Workflows 4.0, a feature enabling users to extend workflow capabilities by integrating custom storage backends for artifacts. Previously, custom storage required manual scripting, leading to loss of core features like syntax for passing data, garbage collection, and UI access. Plugins, implemented as gRPC servers packaged in Docker images, run alongside workflow pods and the Argo UI, providing a seamless experience for managing artifacts in non-native storage. This significantly enhances flexibility for organizations with specialized storage needs, without requiring a fork of Argo or complex CRD changes.

**Technologies:** Argo Workflows, Argo Workflows Artifact Driver Plugins, gRPC, Kubernetes Init Containers, Kubernetes Sidecar Containers

### 9. GitOps and Secrets: State of the Union - Kostis Kapelonis, Octopus Deploy

[https://www.youtube.com/watch?v=6d4dmR7pcPE](https://www.youtube.com/watch?v=6d4dmR7pcPE) · 25m 25s

This talk addresses the challenge of managing secrets within a GitOps workflow using Argo CD. The speaker highlights that directly storing raw secrets in Git is a recipe for disaster and argues against making Argo CD a secret management solution itself. Instead, the presentation strongly recommends leveraging external secret operators like the External Secret Operator (ESO) or Secrets Store CSI Driver, which handle the fetching and processing of secrets from dedicated secret stores (e.g., HashiCorp Vault), allowing applications to auto-reload secrets without requiring manual sync operations or pod restarts in Argo CD. This decouples secret management from GitOps deployment, promoting a cleaner, more secure, and automated approach.

**Technologies:** Argo CD, External Secret Operator (ESO), Secrets Store CSI Driver, HashiCorp Vault, Bitnami Sealed Secrets

### 10. Anatomy of a GitOps Engine: Understanding Your Sync - Alexandre Gaudreault, Intuit

[https://www.youtube.com/watch?v=0yy4R6KfZSA](https://www.youtube.com/watch?v=0yy4R6KfZSA) · 25m 18s

This talk provides a deep dive into the 'Anatomy of a GitOps Engine,' specifically focusing on Argo CD's synchronization process. It explains how Argo CD interacts with its internal GitOps Engine to manage Kubernetes resources, detailing the steps involved in diffing, applying, and syncing. The presentation covers advanced features such as sync hooks (PreSync, Sync, PostSync, SyncFail), pruning resources (including PruneLast and Prune=confirm options), sync waves for ordered deployments, and self-healing mechanisms, offering crucial insights into how these features influence deployment workflows and resource management within a GitOps framework.

**Technologies:** Argo CD, Kubernetes, GitOps Engine, GitOps Promoter

### 11. Empowering Autonomy: BYD's Journey Taming Million-Task Scale With Argo... Shuangkun Tian & Zhang Bao

[https://www.youtube.com/watch?v=tGaHB5uF7XA](https://www.youtube.com/watch?v=tGaHB5uF7XA) · 26m 42s

BYD modernized its autonomous driving data processing pipelines by migrating from Apache Airflow to a multi-cluster Argo Workflows architecture, addressing critical scalability, mutability, and GitOps limitations. This transition enabled the processing of millions of concurrent tasks, achieved 11x faster execution, and reduced compute costs by 30%. The solution integrated Argo Workflows as a supervisor for RayCluster management, utilized native concurrency controls, and leveraged Kubernetes scheduler quotas to enhance stability, efficiency, and observability at extreme scale.

**Technologies:** Argo Workflows, Kubernetes (K8s), Ray, Argo CD

### 12. Beyond Argo Rollouts: Boosting Developer Experience With Intelligen... Carlos Sanchez & Kevin Dubois

[https://www.youtube.com/watch?v=BATVldmFGFs](https://www.youtube.com/watch?v=BATVldmFGFs) · 25m 33s

This talk presents an intelligent AIOps system that integrates with Argo Rollouts to enhance progressive delivery. By deploying AI agents (built using Quarkus and Java) alongside canary deployments, the system can autonomously analyze real-time metrics and logs to detect failures. If an issue is found, the AIOps system can trigger an automatic rollback and either create a GitHub Pull Request with a proposed code fix for simpler problems or generate a detailed GitHub Issue for more complex performance degradations, thereby significantly improving developer experience and deployment safety.

**Technologies:** Argo Rollouts, Prometheus, Kubernetes, Quarkus, GitHub

### 13. Don't Do What Donny Don't Does: The 10 Dos and 500 Don’ts of Workflows... Tim Collins & Becky Pauley

[https://www.youtube.com/watch?v=tzCGRezWJzQ](https://www.youtube.com/watch?v=tzCGRezWJzQ) · 23m 35s

This talk provides essential 'dos and don'ts' for effectively using Argo Workflows, emphasizing that successful implementation, especially at scale, requires a deep understanding of its Kubernetes foundation. It highlights critical aspects such as remembering Argo's architecture as a Kubernetes operator, meticulous observation of cluster health and logs, precise resource allocation for pods, and robust security practices including proper permissions and network access. Adhering to these guidelines helps prevent misconfigurations, optimize performance, reduce cloud costs, and enhance the reliability of workflows.

**Technologies:** Argo Workflows, Kubernetes, Tailscale, Pipekit, Renovate

### 14. Panel: Building a GitOps... Dan Garfield, Revital Barletz, Antonela Cukurin, & Gabriel Quennesson

[https://www.youtube.com/watch?v=EYk72vXAJrU](https://www.youtube.com/watch?v=EYk72vXAJrU) · 37m 15s

This panel delves into the cultural challenges of building a GitOps culture, rather than just technical hurdles. Speakers from Novo Nordisk and Michelin share their experiences in adopting Argo CD for Kubernetes deployments at scale, emphasizing the need for robust guardrails, internal evangelization, and empowering development teams. The discussion highlights the integration of GitOps with existing ITIL processes, navigating compliance in regulated industries, and the future role of GitOps in managing autonomous AI agents through code-driven workflows.

**Technologies:** Argo CD, Kubernetes, GitOps, Argo CD Operator, ITIL (Information Technology Infrastructure Library)

### 15. Panel: From Signals To Safety Nets: Self-Healing Progressive... Anastasiia G, Julia Furst M & Joe F

[https://www.youtube.com/watch?v=VXYU0olNwN8](https://www.youtube.com/watch?v=VXYU0olNwN8) · 27m 56s

This panel discussion at ArgoCon Europe explores how to achieve self-healing progressive delivery using Argo Rollouts and Linkerd. The speakers highlight the 'Deployment Blast Radius Problem,' where frequent small changes lead to more outages, emphasizing that traditional application metrics are often insufficient for detecting subtle degradations. The proposed solution involves a closed-loop system where Argo Rollouts uses Linkerd's network-level observability to automatically monitor golden signals and trigger rollbacks without human intervention, effectively creating 'safety nets' for deployments. A live demo showcased this auto-rollback mechanism in action when a canary deployment's success rate dropped below a defined threshold.

**Technologies:** Argo Rollouts, Linkerd, Prometheus, Kubernetes, Argo CD

### 16. Analysis in Action: Real-World Metrics Driven Argo Roll... Chris Detsicas & Antonio Jimenez Martinez

[https://www.youtube.com/watch?v=Cj5_wyUPhNg](https://www.youtube.com/watch?v=Cj5_wyUPhNg) · 22m 13s

This talk presents how Argo Rollouts, a Kubernetes controller, enables progressive application rollouts with continuous verification driven by metrics and custom analysis. It highlights the limitations of native Kubernetes deployments (RollingUpdate, Recreate) and demonstrates how Argo Rollouts facilitates advanced deployment strategies like Blue/Green and Canary releases. Through `AnalysisTemplates`, engineers can automate promotion and rollback decisions based on real-time metrics, webhooks, or Kubernetes Jobs, ensuring applications behave as expected and enabling safe, efficient, and automated software delivery.

**Technologies:** Argo Rollouts, Kubernetes, AnalysisTemplates, Prometheus, Istio

### 17. GitOps Your Costs: Automated FinOps Through Argo Workflows - Kunal Das & Hrittik Roy

[https://www.youtube.com/watch?v=x5nVyVZSd0A](https://www.youtube.com/watch?v=x5nVyVZSd0A) · 25m 9s

This talk introduces an automated FinOps solution leveraging the Argo ecosystem to combat Kubernetes cost overruns, primarily caused by structural overprovisioning. The proposed closed-loop system uses Argo Workflows for scheduled cost optimization, Argo Events for reactive cost management, Argo CD for GitOps-driven delivery with auditing and rollbacks, and Argo Rollouts for progressive and safe resource changes. This approach aims to bridge the gap between resource observation and proactive action, offering significant cost savings without degrading performance.

**Technologies:** Argo Workflows, Argo Events, Argo CD, Argo Rollouts, Kubecost

### 18. Argo CD: Previewing Pull Request Changes in SECONDS! - Dag Bjerre Andersen & Sergey Shevchenko

[https://www.youtube.com/watch?v=fcajag5di68](https://www.youtube.com/watch?v=fcajag5di68) · 21m 19s

This talk introduces Argo CD Diff Preview, a tool designed to provide fast, reliable, and accurate previews of Kubernetes manifest changes directly within pull requests. It addresses the common GitOps challenge of verifying templated YAML before merging, offering two modes: an ephemeral Kubernetes cluster for complete isolation or a pre-configured Argo CD instance for speed. The tool ensures deterministic diffs by comparing static desired states from Git branches, significantly reducing review times from minutes to seconds and enhancing deployment confidence.

**Technologies:** Argo CD Diff Preview, Argo CD, Kubernetes (K8s), Helm, Kustomize

### 19. From Canary To Global: Unified Progressive Delivery for Hybrid Cloud With... Zhuang Zhang & Ryan Wu

[https://www.youtube.com/watch?v=Eh22awtCH7I](https://www.youtube.com/watch?v=Eh22awtCH7I) · 25m 33s

This talk introduces a unified progressive delivery platform for hybrid cloud environments, addressing challenges in managing Kubernetes across multiple clusters. By integrating ArgoCD and Karmada, the solution abstracts underlying cluster complexities from developers, enabling global release orchestration through "cluster-abstracted waves." Karmada extends Kubernetes APIs for multi-cluster management, offering robust scheduling policies, resource interpretation, and built-in failover capabilities. This synergy provides advanced features like workload division, cross-cluster customization, and region-level canary releases, ensuring safe, efficient, and highly available application deployments.

**Technologies:** Argo Rollouts, Argo ApplicationSet, Karmada, Apache ZooKeeper (ZK), Colima

### 20. Decoupling Canary Deployments From DBs With Argo Rollouts - Anastasiia Gubska, Chase UK

[https://www.youtube.com/watch?v=6xZdcPQL3xs](https://www.youtube.com/watch?v=6xZdcPQL3xs) · 27m 11s

This talk presents strategies for decoupling canary deployments from database schema changes using Argo Rollouts. It addresses the inherent challenges of progressive delivery with stateful applications, where tightly coupled application and database changes can lead to data loss during rollbacks or failed deployments. The speaker demonstrates how adopting backward and forward-compatible database refactoring patterns, combined with multi-stage rollouts orchestrated by Argo Rollouts, enables zero-downtime database schema evolution and reliable application rollbacks, ensuring a safer and more robust deployment pipeline.

**Technologies:** Argo Rollouts, Kubernetes, PostgreSQL, Git/GitHub, CNCF (Cloud Native Computing Foundation)

### 21. Network Segmentation at Scale: GitOps-Driven Multi-Tenant Isol... Luke Philips & Christian Hernandez

[https://www.youtube.com/watch?v=BPbv78VHYbQ](https://www.youtube.com/watch?v=BPbv78VHYbQ) · 25m 5s

This talk demonstrates how to achieve robust network segmentation and multi-tenant isolation in Kubernetes using a GitOps-driven approach with Argo CD and Cilium. It highlights that Kubernetes namespaces alone are insufficient for security boundaries and showcases how to enforce L3-L7 traffic filtering with Cilium Network Policies, establish orchestration boundaries and tenant self-service with Argo CD Projects and RBAC, and automate tenant onboarding at scale using Argo CD ApplicationSets. The solution emphasizes self-healing capabilities and real-time observability for compliance and security.

**Technologies:** Kubernetes, Argo CD, Cilium, Hubble, Argo CD ApplicationSets

### 22. Beyond GitOps: Building Intelligent Drift Detection and... Ram Mohan Rao Chukka & Shibi Ramachandran

[https://www.youtube.com/watch?v=TEYa2MOhvRU](https://www.youtube.com/watch?v=TEYa2MOhvRU) · 19m 42s

This talk presents an evolution of GitOps within Argo CD to address limitations in handling operational drift, particularly when manual changes are legitimate. It introduces a two-stage approach: initially integrating Open Policy Agent (OPA) for policy-driven decisions (ALLOW, DEFER, ENFORCE) to add nuance beyond a binary sync. The second, AI-powered stage, incorporates an AI agent using Model Context Protocol (MCP) and Agent-to-Agent (A2A) communication to query external monitoring and incident management tools, enabling context-aware decisions and auto-remediation that prevents unnecessary rollbacks and reduces manual intervention.

**Technologies:** Argo CD, Open Policy Agent (OPA), Model Context Protocol (MCP), Agent-to-Agent (A2A), Prometheus

### 23. Cross-Cluster Progressive Delivery: Rolling Out Apps With Arg... Carlos Santana & Elamaran Shanmugam

[https://www.youtube.com/watch?v=_2vRm81KyIk](https://www.youtube.com/watch?v=_2vRm81KyIk) · 28m 10s

This talk presents a solution for cross-cluster progressive delivery using the Argo ecosystem, enhanced by Hera and AI. It addresses the complexity of deploying applications across multiple Kubernetes clusters by orchestrating Argo Rollouts (for intra-cluster canary deployments) with Argo Workflows (for inter-cluster staged rollouts). The speakers introduce Hera, a Python SDK for Argo Workflows, enabling pipeline definition in Python instead of YAML. They demonstrate an AI-powered custom UI that generates these Hera promotion pipelines from natural language prompts, offering a streamlined "single pane of glass" for monitoring and controlling cross-cluster deployments.

**Technologies:** Argo Rollouts, Argo Workflows, Argo Events, ArgoCD, Hera

### 24. Agnostic Workload Identity Across the Argo Ecosystem Using SPIFFE... Andrew Block & Blake Pettersson

[https://www.youtube.com/watch?v=rTH0NBaWnCI](https://www.youtube.com/watch?v=rTH0NBaWnCI) · 28m 50s

This talk addresses the challenges of managing long-lived secrets and inconsistent credential rotation within the Argo ecosystem by advocating for agnostic workload identity using SPIFFE and SPIRE. The speakers highlight how workload identity offers short-lived, automatically rotated, and fine-grained cryptographic identities for software, contrasting it with traditional human-centric authentication. While Argo CD currently has limited and inconsistent workload identity support, a proposed initiative aims to integrate native SPIFFE-based identity for repository credentials, offering a simpler, more secure, and dependency-free solution with in-memory tokens and per-project access.

**Technologies:** SPIFFE, SPIRE, Argo CD, Kubernetes Secrets, OIDC

### 25. From FTP To Argo CD & Argo Rollouts: An Adoption History Insid... Alfonso Ming & Meritxell Rodríguez

[https://www.youtube.com/watch?v=Wrek7kjrcs0](https://www.youtube.com/watch?v=Wrek7kjrcs0) · 21m 58s

This talk by Alfonso Ming and Meritxell Rodríguez from LIDL HUG DIGITAL details their journey from anxiety-inducing manual deployments and ClickOps to a robust, automated GitOps model using Argo CD and Argo Rollouts. They transitioned from sequential, manual steps across multiple environments, which often led to costly production bugs and lengthy rollbacks, to a declarative system where Git is the single source of truth. This transformation enabled rapid, safe deployments with canary releases and automated rollbacks, drastically reducing recovery times and fostering operational peace of mind for their teams managing over 100 million users.

**Technologies:** Argo CD, Argo Rollouts, Kubernetes, GitOps, Terraform

### 26. Declarative...ish? Fixing Hidden Argo CD Pitfalls in Your GitOps Setup - Regina Voloshin

[https://www.youtube.com/watch?v=VfVQj4Oa3z0](https://www.youtube.com/watch?v=VfVQj4Oa3z0) · 26m 46s

This talk identifies several hidden anti-patterns in Argo CD GitOps setups that lead to complexity and hinder scalability. Key pitfalls include hard-coding Helm or Kustomize data directly into Argo CD Application resources, imperatively creating applications, and packaging applications within nested Helm charts (the 'Helm Sandwich'). The overarching recommendation is to adopt a 'more GitOpsy' approach by utilizing ApplicationSets, ensuring application configurations exclusively reference external files in Git, and storing all Application manifests declaratively in Git to maintain a single source of truth and simplify development workflows.

**Technologies:** Argo CD, Helm, Kustomize, ApplicationSet, Git

### 27. Leveraging Argo Workflows To Move, Migrate and Manage All of Your Workflows - John Keates

[https://www.youtube.com/watch?v=H7cayhuvnD8](https://www.youtube.com/watch?v=H7cayhuvnD8) · 27m 15s

This talk by John Keates from Wehkamp Retail Group addresses the challenge of managing diverse and often fragmented workflows across an e-commerce platform, ranging from traditional IT service desk requests to complex CI/CD pipelines and data engineering tasks. Keates highlights the 'tribal knowledge' burden and operational inefficiencies arising from disparate systems and programming languages. The solution proposed is to consolidate and standardize these workflows using the Argo ecosystem, specifically leveraging Argo Workflows for execution and Argo Events for event-driven orchestration, enabling a more unified, observable, and maintainable platform.

**Technologies:** Argo Workflows, Argo Events, Argo CD, Kubernetes, Jenkins

### 28. Lightning Talk: The $10,000 Argo CD Mistake: Eliminating Phantom Sync... Vanshika Jain & Aditya Soni

[https://www.youtube.com/watch?v=Gc12_8tx3zg](https://www.youtube.com/watch?v=Gc12_8tx3zg) · 9m 9s

This lightning talk addresses the significant performance and cost issues, dubbed the "$10,000 Argo CD Mistake," caused by inefficient Argo CD Repo Server scaling and "phantom syncs." The speakers explain how default aggressive polling of Git repositories, especially with hundreds of applications, leads to high CPU/memory consumption and slow syncs, even when no changes occur. They present four key fixes: reducing polling frequency, implementing event-driven webhooks, utilizing Argo CD ApplicationSet for automated manifest generation, and optimizing Git fetches with shallow clones. These optimizations collectively achieved substantial improvements in CPU/memory usage and sync times, emphasizing architectural flow over mere code changes.

**Technologies:** Argo CD, Kubernetes, Argo CD ApplicationSet, Webhooks, Git

### 29. Lightning Talk: “Tilt? Goreman? Make? Go Run? A Contributor’s Guide To Argo CD Local... Nitish Kumar

[https://www.youtube.com/watch?v=XW--BOC1s_Y](https://www.youtube.com/watch?v=XW--BOC1s_Y) · 6m 23s

This lightning talk addresses the complexities of local development for Argo CD, a distributed cloud-native project, highlighting the challenges of managing multiple components and build processes. The speaker, an Argo CD maintainer, introduces `make targets` and `Goreman` as traditional methods but advocates for `Tilt` as a superior solution. `Tilt` automates the build, deployment, and live updating of Argo CD components within a local Kubernetes cluster, significantly streamlining the developer workflow for both frontend and backend changes.

**Technologies:** Argo CD, Tilt, Goreman, Kubernetes

### 30. ArgoCon | Closing Remarks - Christian Hernandez, Program Committee Co-Chair

[https://www.youtube.com/watch?v=xnRBDiOSQv8](https://www.youtube.com/watch?v=xnRBDiOSQv8) · 1m 14s

This session features Christian Hernandez delivering the closing remarks for ArgoCon Europe, thanking attendees and encouraging participation in an extensive community survey. The survey aims to gather feedback from Argo project users to enhance future events and community development. Hernandez also announced that the next ArgoCon will take place in Barcelona.

**Technologies:** Argo, Argo CD

### 31. ArgoCon | Welcome + Opening Remarks - Dan Garfield, Program Committee Co-Chair

[https://www.youtube.com/watch?v=LbAm7UnhDHY](https://www.youtube.com/watch?v=LbAm7UnhDHY) · 11m 6s

This talk provides a welcome and opening remarks for ArgoCon Europe, focusing on the "Argo Project Velocity Update." Dan Garfield, an Argo Maintainer, highlights the diverse set of Argo projects including Workflows, CD, Rollouts, and Events, emphasizing their growing importance in cloud-native and AI/data-driven environments. He showcases significant community growth, with over 2,000 contributors and a high number of maintainers, underscoring the project's robust activity and corporate support. Case studies demonstrate substantial cost reductions and faster upgrade times achieved with Argo CD, while new 'Argo Project Labs' explore innovations like GitOps promotion controllers and AI assistants.

**Technologies:** Argo Workflows, Argo CD, Argo Rollouts, Argo Events, Kubernetes

