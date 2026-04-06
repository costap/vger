# Playlist Digest — PL2yQDdvlhXf9l1y_VZAm0vG2hproZMtip

*30 talks analysed by V'Ger*

## Overarching Theme

This playlist highlights the evolving landscape of cloud-native data management on AWS, emphasizing the critical role of highly scalable, resilient, and cost-optimized databases for modern microservices architectures. A dominant theme is the deep integration of Generative AI capabilities across various data stores—from relational and NoSQL to graph and caching—leveraging vector embeddings and knowledge graphs to create intelligent, context-aware applications while addressing foundational concerns like data consistency, performance, and operational efficiency.

## Watch These First

1. **[DraftKings: Scaling to 1M operations/min on Aurora during Super Bowl (DAT320)](https://www.youtube.com/watch?v=sx3mvBQtaX8)**  
   This talk offers highly relevant, real-world insights for a sportsbook platform running on EKS, demonstrating how to achieve extreme scalability with Aurora MySQL and sharding to handle massive peak loads effectively.

2. **[Advanced data modeling for Amazon DynamoDB (DAT414)](https://www.youtube.com/watch?v=PGfyJX8eS48)**  
   DynamoDB is crucial for high-performance microservices. This session provides essential advanced data modeling techniques and anti-patterns, which are critical for optimizing cost, performance, and efficiency in a distributed environment.

3. **[A practitioner’s guide to data for agentic AI (DAT315)](https://www.youtube.com/watch?v=XLWjq5FInyQ)**  
   This talk provides a foundational understanding of how to architect data for modern AI agents, covering key concepts like the ReAct loop, memory components, caching, and the Model Context Protocol (MCP), essential for building intelligent applications.

4. **[Optimize agentic AI apps with semantic caching in Amazon ElastiCache (DAT451)](https://www.youtube.com/watch?v=BCr53N1oVX0)**  
   Semantic caching with ElastiCache is a critical optimization for AI applications, directly addressing compounding latency and cost at scale. This is highly relevant for improving performance and cost-efficiency in AI-powered microservices.

5. **[Deep dive into Amazon Aurora DSQL and its architecture (DAT439)](https://www.youtube.com/watch?v=sCHlFiEEY6E)**  
   Aurora DSQL represents the future of distributed, serverless PostgreSQL. Understanding its architecture for active-active writes and strong consistency offers forward-looking insights for achieving ultra-high scale and resilience for transactional workloads.

## Recommended Learning Path

1. Amazon Aurora (PostgreSQL/MySQL)
2. Amazon DynamoDB
3. Amazon ElastiCache (Valkey/Redis)
4. Amazon Bedrock
5. Vector Embeddings (pgvector, HNSW)
6. Model Context Protocol (MCP)
7. Amazon Neptune
8. Amazon Aurora DSQL
9. AWS Zero-ETL / AWS Database Migration Service (DMS)

## Key Insights

For an engineer operating a microservices sportsbook on AWS EKS with Istio, the key takeaways from this playlist are multi-faceted. First, databases are no longer just passive storage but active components in Generative AI architectures; leveraging existing data stores with vector embeddings, semantic caching, and knowledge graphs (like Amazon Neptune) is crucial for building differentiated and intelligent applications. Second, achieving extreme scalability, high resilience, and cost optimization remains paramount. This requires deep understanding of database-specific data modeling (e.g., DynamoDB), advanced HA/DR patterns (Aurora Global Database, Aurora DSQL), efficient caching strategies (ElastiCache Serverless), and cost management tools like AWS Database Savings Plans. Third, integrating LLMs into existing microservices platforms demands new secure and efficient interaction patterns, with the Model Context Protocol (MCP) emerging as a critical approach for LLMs to safely query structured data. Finally, the playlist emphasizes the growing importance of hybrid AI architectures that combine the strengths of symbolic AI (for reasoning and explainability) with LLMs, alongside streamlined data movement and modernization through AWS Zero-ETL integrations and DMS.

## Technology Radar

| Technology | Talks |
|---|---|
| Amazon DynamoDB | 10 |
| PostgreSQL | 9 |
| Amazon Bedrock | 8 |
| Amazon Aurora | 6 |
| Amazon ElastiCache | 6 |
| AWS Lambda | 5 |
| Amazon Neptune | 4 |
| Amazon OpenSearch Service | 4 |
| Amazon RDS | 4 |
| Amazon S3 | 4 |
| Valkey | 4 |
| AWS Database Migration Service (DMS) | 3 |
| AWS Glue | 3 |
| Amazon Aurora DSQL | 3 |
| Amazon DocumentDB | 3 |
| Amazon Redshift | 3 |
| MySQL | 3 |
| Amazon Aurora PostgreSQL | 2 |
| Amazon EKS | 2 |
| Amazon ElastiCache Serverless | 2 |

## All Talks

### 1. AWS re:Invent 2025 - Optimize agentic AI apps with semantic caching in Amazon ElastiCache (DAT451)

[https://www.youtube.com/watch?v=BCr53N1oVX0](https://www.youtube.com/watch?v=BCr53N1oVX0) · 43m 51s

This talk introduces semantic caching as a crucial optimization technique for agentic AI applications, addressing challenges like compounding latency and cost at scale. It explains how traditional caching falls short for the semantic nature of AI queries and details the mechanics of semantic caching, which uses embedding models to convert queries into vectors and then performs approximate nearest neighbor searches (like HNSW) to find semantically similar, previously computed responses. The speakers highlight the benefits of this approach in reducing LLM calls, costs, and latency while maintaining high accuracy, particularly with the new vector search capabilities in Amazon ElastiCache for Valkey.

**Technologies:** Amazon ElastiCache, Valkey, Amazon Bedrock

### 2. AWS re:Invent 2025 - Web3 + AI: Agents that mean business (DAT318)

[https://www.youtube.com/watch?v=wuB3Sq5AWo8](https://www.youtube.com/watch?v=wuB3Sq5AWo8) · 56m 29s

This talk explores the intersection of Web3 and AI, focusing on how agentic payments can revolutionize financial transactions. It delves into the underlying decentralized technologies, AWS's perspective on Web3, and how Coinbase is leveraging open standards like X402 to build scalable payment solutions. The speakers highlight the benefits of Web3 payments in addressing challenges like high processing fees, settlement delays, and currency fluctuations, ultimately enabling micro-payments and banking the unbanked.

**Technologies:** X402, Amazon Bedrock, AWS Nitro Enclaves

### 3. AWS re:Invent 2025 - A tale of two transactions (DAT455)

[https://www.youtube.com/watch?v=SNnUpYvBfow](https://www.youtube.com/watch?v=SNnUpYvBfow) · 57m 53s

This talk explores the complex tradeoffs involved in selecting database isolation levels, moving beyond the common misconception of a single 'best' level. It delves into classical algorithms like Two-Phase Locking (2PL) and Optimistic Concurrency Control (OCC), highlighting their implications for correctness, concurrency, throughput, and high availability. The speaker emphasizes how multi-version concurrency control (MVCC), as implemented in Aurora D-SQL and PostgreSQL, offers a more efficient approach to isolation, particularly for distributed systems, by allowing concurrent reads without blocking writers and vice versa. Ultimately, the choice of isolation level is presented as a non-linear decision heavily dependent on specific workload patterns, with strong consistency often being the optimal, yet operationally challenging, choice for cloud-native applications.

**Technologies:** EC2, EBS, Lambda, Amazon Aurora (DSQL), PostgreSQL

### 4. AWS re:Invent 2025 - Supercharge app intelligence using gen AI with Amazon DocumentDB (DAT313)

[https://www.youtube.com/watch?v=YL1eKFuwFqw](https://www.youtube.com/watch?v=YL1eKFuwFqw) · 54m 8s

This talk explores how to supercharge application intelligence using generative AI with Amazon DocumentDB. It highlights that an organization's proprietary data is the key differentiator for AI applications, emphasizing the importance of leveraging existing data sources and architectures. The speakers demonstrate how DocumentDB supports vector embeddings for semantic search, enabling the creation of intelligent chatbots and data analysis tools. They introduce a TSQL plugin for mongosh and showcase a Retrieval Augmented Generation (RAG) architecture built with DocumentDB, LangChain, and Bedrock, demonstrating its multilingual capabilities and schema generation features.

**Technologies:** Amazon DocumentDB, Vector Embeddings, Amazon Bedrock, LangChain, Gradio

### 5. AWS re:Invent 2025 - Deep Dive into Deloitte's Amazon Neptune GenAI Security Intelligence Center

[https://www.youtube.com/watch?v=KD5C93kwBMg](https://www.youtube.com/watch?v=KD5C93kwBMg) · 54m 41s

This talk presents Deloitte's Amazon Neptune GenAI Security Intelligence Center, which leverages graph-RAG techniques to address information overload and context problems faced by SecOps engineers. The solution utilizes the open-source GraphRAG Toolkit to build a Hierarchical Lexical Graph from unstructured data, enabling contextual analysis for security alerts. This approach aims to provide reasoned, explainable, and human-augmented decision-making, moving towards a 'zero disconnect' cyber defense by integrating short-term operational data with long-term organizational knowledge.

**Technologies:** Amazon Neptune, GraphRAG Toolkit, LlamaIndex, Wiz, CrowdStrike

### 6. AWS re:Invent 2025 - Build serverless chatbots using Amazon ElastiCache & Aurora PostgreSQL (DAT314)

[https://www.youtube.com/watch?v=eCiR1mDe5Rw](https://www.youtube.com/watch?v=eCiR1mDe5Rw) · 41m 15s

This talk presents an architecture for building performant, serverless chatbots using Amazon ElastiCache Serverless for caching and Amazon Aurora PostgreSQL with the pgvector extension for semantic search. The core problem addressed is slow chatbot response times leading to high customer abandonment. The solution leverages tiered memory management and multi-agent workflows, enabling personalized and context-aware interactions while ensuring scalability, resilience, and cost optimization.

**Technologies:** Amazon ElastiCache Serverless, Amazon Aurora PostgreSQL, pgvector, Amazon Bedrock, Strands Agents

### 7. AWS re:Invent 2025 - Symbolic AI in the age of LLMs (DAT443)

[https://www.youtube.com/watch?v=Atf4DVKGuMg](https://www.youtube.com/watch?v=Atf4DVKGuMg) · 53m 51s

This talk explores the history and current relevance of Symbolic AI, particularly in the context of Large Language Models (LLMs). Dr. Lassila highlights Symbolic AI's long-standing principles, such as knowledge graphs and ontologies, for representing explicit knowledge and enabling logical reasoning. He discusses the strengths and weaknesses of both symbolic and non-symbolic AI approaches, emphasizing how hybrid architectures combining these paradigms can mitigate challenges like hallucinations and improve explainability and computational efficiency in generative AI systems. The presentation advocates for leveraging mature symbolic AI techniques alongside modern LLMs to build more robust and trustworthy intelligent systems.

**Technologies:** RDF, SPARQL, OWL, SHACL, Amazon Neptune

### 8. AWS re:Invent 2025 - Cut costs & operate efficiently on Amazon RDS for SQL Server & Oracle (DAT325)

[https://www.youtube.com/watch?v=06l39fx_AXY](https://www.youtube.com/watch?v=06l39fx_AXY) · 38m 51s

This talk highlights strategies to reduce costs and enhance operational efficiency for SQL Server and Oracle workloads on Amazon RDS. It introduces new capabilities like additional storage volumes for increased flexibility and throughput, support for SQL Server Developer Edition to lower licensing costs for dev/test environments, and Multi-AZ high availability for SQL Server Web Edition. The session also covers CPU optimization for 7th generation instances and the cost benefits of using bare metal instances for Oracle databases.

**Technologies:** Amazon RDS, Microsoft SQL Server Developer Edition, Microsoft SQL Server Web Edition, Oracle Database on Amazon RDS, Amazon EBS

### 9. AWS re:Invent 2025 - Deep dive into databases zero-ETL integrations (DAT445)

[https://www.youtube.com/watch?v=rwSfnVd1Bgs](https://www.youtube.com/watch?v=rwSfnVd1Bgs) · 35m 48s

The talk provides a deep dive into AWS Zero-ETL integrations, a fully managed service designed to simplify data movement for analytical workloads. It addresses common data engineering challenges like complex data pipelines and operational analytics by enabling near real-time data synchronization from transactional databases like Amazon Aurora and DynamoDB to analytical stores such as Amazon Redshift, Amazon S3 data lakes, and Amazon OpenSearch Service. The speaker illustrates various use cases, including sentiment analysis with GenAI, and highlights how Zero-ETL automates schema changes and data updates, reducing operational overhead and allowing engineers to focus on higher-value tasks.

**Technologies:** AWS Zero-ETL, Amazon Aurora, Amazon DynamoDB, Amazon Redshift, Amazon OpenSearch Service

### 10. AWS re:Invent 2025 - Amazon Aurora HA and DR design patterns for global resilience (DAT442)

[https://www.youtube.com/watch?v=TNURufOhu_A](https://www.youtube.com/watch?v=TNURufOhu_A) · 45m 55s

This talk explores Amazon Aurora's high availability (HA) and disaster recovery (DR) design patterns to achieve global resilience. It covers durability with six data copies across three Availability Zones (AZs) and zero RPO, multi-AZ instance availability with rapid failover, and scaling capabilities through read replicas and automatic storage scaling. The speakers also detail advanced connection management, Aurora Global Database for cross-region replication and failover, and streamlined maintenance processes including automated upgrades and Blue/Green deployments. The talk concludes with an introduction to Aurora DSQL, a distributed PostgreSQL offering active-active writes and fully automated resilience across regions.

**Technologies:** Amazon Aurora, PostgreSQL, MySQL, Amazon S3, Amazon Route 53

### 11. AWS re:Invent 2025 - How Samsung optimized 1.2 PB on Amazon DynamoDB with zero downtime (DAT323)

[https://www.youtube.com/watch?v=Q_OJuN208gw](https://www.youtube.com/watch?v=Q_OJuN208gw) · 21m 58s

Samsung Cloud faced a massive data growth challenge with its 1.2 PB tab data table on Amazon DynamoDB, incurring significant storage costs. Through data analysis, they discovered 90% of the data were old deletion records (tombstones). By deeply understanding user behavior and leveraging the Samsung Internet's 99-tab limit policy, they re-evaluated their data retention strategy. This led to a migration strategy that reduced the data volume to be migrated by 90% and ensured zero downtime for users, exceeding cost-saving goals by 150% and completing the migration in just one week.

**Technologies:** Amazon DynamoDB, Amazon ElastiCache, Local Secondary Index (LSI)

### 12. AWS re:Invent 2025 - Multi-Region strong consistency with Amazon DynamoDB global tables (DAT440)

[https://www.youtube.com/watch?v=m6V5DsoU7X4](https://www.youtube.com/watch?v=m6V5DsoU7X4) · 50m 50s

This talk explores multi-region strong consistency with Amazon DynamoDB global tables, focusing on the two primary replication modes: asynchronous and multi-region strong consistency (MRSC). It highlights the challenges of multi-region data, such as latency, write ordering, and conflict resolution, and discusses the trade-offs between predictable performance and strict consistency. The speakers detail how DynamoDB global tables are designed to provide high availability and business continuity, enabling architects to choose the appropriate consistency model (eventual or strong) based on their specific workload requirements and recovery objectives.

**Technologies:** Amazon DynamoDB Global Tables, Asynchronous Global Tables, Multi-Region Strong Consistency (MRSC) Global Tables

### 13. AWS re:Invent 2025 - Using generative AI to accelerate database modernization with AWS DMS (DAT420)

[https://www.youtube.com/watch?v=sogq2AHaqJc](https://www.youtube.com/watch?v=sogq2AHaqJc) · 54m 36s

The talk introduces AWS's generative AI capabilities within AWS Database Migration Service (DMS) and AWS Transform to accelerate database and application modernization. It demonstrates how AI can automate complex schema and code conversions from proprietary databases like Sybase and SQL Server to open-source PostgreSQL, including dependent .NET applications. The speakers highlight the benefits of using AI to overcome common migration challenges, reducing manual effort and time.

**Technologies:** AWS Database Migration Service (DMS), AWS Transform, Amazon Bedrock, PostgreSQL, Sybase (SAP Sybase ASE)

### 14. AWS re:Invent 2025 - A practitioner’s guide to data for agentic AI (DAT315)

[https://www.youtube.com/watch?v=XLWjq5FInyQ](https://www.youtube.com/watch?v=XLWjq5FInyQ) · 58m 30s

This talk presents a practitioner's guide to data for agentic AI, tracing the evolution of AI from simple chatbots with vector search to autonomous AI agents capable of reasoning, planning, and task execution. It introduces the foundational 'ReAct' loop (Reason, Act, Remember) and explores memory components, caching strategies, and data governance for agentic AI applications. The Model Context Protocol (MCP) is highlighted as a standardized way for agents to interact with diverse data sources via tools, enabling efficient and secure data access for both external customers and internal users like data engineers and claims investigators.

**Technologies:** Amazon Bedrock, AWS Lake Formation, Amazon SageMaker, AWS Glue, Amazon ElastiCache

### 15. AWS re:Invent 2025 - DraftKings: Scaling to 1M operations/min on Aurora during Super Bowl (DAT320)

[https://www.youtube.com/watch?v=sx3mvBQtaX8](https://www.youtube.com/watch?v=sx3mvBQtaX8) · 18m 25s

This talk details how DraftKings scaled its financial ledger to handle over 1 million operations per minute, particularly during high-traffic events like the Super Bowl. The architecture leverages Amazon Aurora MySQL, employing sharding to enhance throughput, reduce costs, and accelerate data warehousing. Key to their success was separating reads from writes using read replicas and optimizing stored procedures.

**Technologies:** Amazon Aurora MySQL, Kubernetes, Amazon EKS, CloudWatch Database Insights, Message Broker

### 16. AWS re:Invent 2025 - Revealing the northern lights: Amazon Aurora security deep dive (DAT456)

[https://www.youtube.com/watch?v=yk1m5DqBg9M](https://www.youtube.com/watch?v=yk1m5DqBg9M) · 44m 7s

This talk provides a deep dive into the security architecture of Amazon Aurora, highlighting its cloud-native design that separates compute from storage. It contrasts Aurora with traditional single-host relational databases (like Amazon RDS) and highly scalable NoSQL databases (like Amazon DynamoDB), positioning Aurora as a cloud-native relational database that offers enhanced availability and durability while maintaining compatibility with open-source engines like MySQL and PostgreSQL. The speakers elaborate on Aurora's layered defense mechanisms, including SELinux, Chronicle telemetry, and VPC Flow Logs, demonstrating how these controls effectively mitigated a zero-day vulnerability (CVE-2024-10979) in PostgreSQL, even without an upstream patch.

**Technologies:** Amazon RDS, PostgreSQL, MariaDB, MySQL, Amazon DynamoDB

### 17. AWS re:Invent 2025 - Advanced data modeling for Amazon ElastiCache (DAT438)

[https://www.youtube.com/watch?v=lwRdG35pA2g](https://www.youtube.com/watch?v=lwRdG35pA2g) · 51m 42s

This talk explores advanced data modeling techniques for Amazon ElastiCache, highlighting its role in building highly scalable, low-latency applications like MMORPGs. It delves into caching strategies such as lazy loading with invalidation and early recaching using TTL to optimize performance and manage database pressure. The session also introduces Valkey as a high-performance key-value store, discussing its data structures like hashes and JSON, and advanced use cases including semantic caching with vector embeddings, publish/subscribe patterns, probabilistic data structures (HyperLogLog, Bloom filter), and geospatial queries for efficient data management.

**Technologies:** Amazon ElastiCache, Valkey, Redis OSS, Memcached, Amazon RDS

### 18. AWS re:Invent 2025 - Advanced data modeling with Amazon DynamoDB (DAT414)

[https://www.youtube.com/watch?v=PGfyJX8eS48](https://www.youtube.com/watch?v=PGfyJX8eS48) · 57m 41s

This talk provides advanced data modeling techniques for Amazon DynamoDB, focusing on its unique characteristics as a fully managed, consumption-based, and consistently performant database. Key concepts include understanding partitioning and the DynamoDB API, efficient use of secondary indexes (especially new multi-attribute composite keys), and strategies for schema evolution. The speaker emphasizes designing data models around access patterns and avoiding common anti-patterns to optimize cost and performance.

**Technologies:** Amazon DynamoDB, DynamoDB Streams, PartiQL, AWS Glue, AWS S3

### 19. AWS re:Invent 2025 - [NEW LAUNCH] Deep dive into AWS Database Savings Plans (DAT326)

[https://www.youtube.com/watch?v=5tT38P-gWA0](https://www.youtube.com/watch?v=5tT38P-gWA0) · 48m 51s

This AWS re:Invent 2025 session introduces the newly launched AWS Database Savings Plans, a highly requested feature offering significant cost savings for database usage. The talk highlights the flexibility of these plans, which cover various database engines, instance types, and deployment options across AWS regions, unlike traditional Reserved Instances. Speakers demonstrate tools like Savings Plans Recommendations and Purchase Analyzer, emphasizing best practices for maximizing savings and facilitating migration from Reserved Instances.

**Technologies:** AWS Database Savings Plans, Amazon RDS, Amazon Aurora, Amazon DynamoDB, Amazon ElastiCache

### 20. AWS re:Invent 2025 - Coding an MCP server for Amazon Aurora  (DAT429)

[https://www.youtube.com/watch?v=H6_YsMnFZRE](https://www.youtube.com/watch?v=H6_YsMnFZRE) · 55m 41s

This talk demonstrates how to code a Model Context Protocol (MCP) server for Amazon Aurora PostgreSQL, enabling large language models (LLMs) to securely interact with relational databases. The speakers build an 'Entertainment Hub' application using Streamlit and Strands, which leverages Amazon Bedrock as the LLM. They showcase how to expose PostgreSQL schema and data via MCP tools, emphasizing the importance of detailed schema comments and row-level security for production environments and mitigating prompt injection risks.

**Technologies:** Model Context Protocol (MCP), Amazon Aurora, Amazon Bedrock, PostgreSQL, Streamlit

### 21. AWS re:Invent 2025 - Deep dive into Amazon Aurora and its innovations (DAT441)

[https://www.youtube.com/watch?v=sfZUQ6sFpxE](https://www.youtube.com/watch?v=sfZUQ6sFpxE) · 56m 5s

This talk provides a deep dive into Amazon Aurora, a cloud-native relational database service, highlighting its architectural innovations and recent updates. Key aspects covered include Aurora's unique storage architecture providing high availability and fault tolerance, new features like local and global write forwarding, and optimized reads through tiered caching. The session also introduces new management capabilities such as fleet-wide upgrade orchestration and zero-ETL integration with other AWS analytics services, demonstrating how Aurora continues to evolve for performance, scalability, and operational simplicity.

**Technologies:** Amazon Aurora, Aurora DSQL, MySQL, PostgreSQL, AWS Lambda

### 22. AWS re:Invent 2025 - Scaling MONOPOLY GO! to millions with DynamoDB & ElastiCache (DAT324)

[https://www.youtube.com/watch?v=P_C6Hv4onuU](https://www.youtube.com/watch?v=P_C6Hv4onuU) · 18m 26s

This talk details how Scopely scaled their mobile game, MONOPOLY GO!, to millions of concurrent users using Amazon DynamoDB and ElastiCache. They highlight the challenges of managing sudden, massive load spikes inherent to live-service mobile gaming, and how DynamoDB's on-demand provisioning with a pre-warm API, and ElastiCache Serverless for auto-scaling, address these issues. The speakers also cover strategies for cost management by optimizing DynamoDB item sizes and the benefits of operational simplicity provided by serverless solutions.

**Technologies:** Amazon DynamoDB, Amazon ElastiCache, Unity, Valkey, Redis

### 23. AWS re:Invent 2025 - Deep dive into Amazon Aurora DSQL and its architecture (DAT439)

[https://www.youtube.com/watch?v=sCHlFiEEY6E](https://www.youtube.com/watch?v=sCHlFiEEY6E) · 45m 48s

Amazon Aurora DSQL is a fully serverless, distributed PostgreSQL database optimized for transactional workloads, contrasting traditional relational databases with horizontally scalable, fault-tolerant distributed systems. The talk dives into DSQL's architecture, detailing the Query Processor, distributed Journal for durability across Availability Zones (and regions), and custom Storage service. It highlights DSQL's active-active, strongly consistent nature achieved through time-based synchronization and optimistic concurrency control, along with transparent automatic scaling via microVMs (Firecracker) and automatic sharding. Connections are secured end-to-end with with AWS IAM integration, simplifying developer experience by abstracting complex backend operations and offering activity-based pricing.

**Technologies:** Amazon Aurora DSQL, PostgreSQL, Firecracker

### 24. AWS re:Invent 2025 - Dive deep into Amazon DynamoDB (DAT435)

[https://www.youtube.com/watch?v=B05YpQ089w8](https://www.youtube.com/watch?v=B05YpQ089w8) · 40m 37s

This talk dives deep into Amazon DynamoDB, highlighting its predictable low latency at any scale as a document and key-value store. The speaker illustrates common customer challenges with DynamoDB through two case studies: unexpected write throttling due to non-random hash key distribution and bizarre latency patterns linked to connection pooling and caching. Practical solutions are provided, emphasizing proper data modeling using multi-attribute keys for global secondary indexes and optimizing client-side connection management to achieve desired performance.

**Technologies:** Amazon DynamoDB, AWS CloudWatch Contributor Insights

### 25. AWS re:Invent 2025 - Deep dive into Amazon DocumentDB and its innovations (DAT444)

[https://www.youtube.com/watch?v=M75JFvwhjtQ](https://www.youtube.com/watch?v=M75JFvwhjtQ) · 45m 20s

This session provides a deep dive into recent innovations and features within Amazon DocumentDB in 2025. Key updates include enhanced compression using Zstandard and LZ4, the introduction of Graviton4 instances for improved performance, and a fully serverless deployment option with fine-grained scaling. The talk also highlights significant improvements to the new query planner for optimized query execution and expanded MongoDB compatibility, all aimed at building high-performance, cost-effective applications.

**Technologies:** Amazon DocumentDB, LZ4, Zstandard, Graviton4, MongoDB API

### 26. AWS re:Invent 2025 - Using graphs over your data lake to power generative AI applications (DAT447)

[https://www.youtube.com/watch?v=pNOn9DHlArQ](https://www.youtube.com/watch?v=pNOn9DHlArQ) · 41m 0s

This talk explores how graph technology can enhance generative AI applications by leveraging relationships within data lakes. Speakers from Amazon Neptune, Brad Bebee and Ümit V. Çatalyürek, discuss strategies for using data lake information in graph models, applying graph algorithms like PageRank and Closeness Centrality for insights, and combining vector search with graph queries for hybrid search. The presentation also introduces GraphRAG, which automatically generates lexical graphs from unstructured data, and persistent memory solutions for AI agents using frameworks like Strands SDK and Bedrock AgentCore, all aimed at improving AI accuracy and explainability.

**Technologies:** Amazon Neptune, Wiz, openCypher, Amazon S3, PageRank

### 27. AWS re:Invent 2025 - Optimizing for high performance with Amazon ElastiCache Serverless (DAT437)

[https://www.youtube.com/watch?v=fXB6wftmigU](https://www.youtube.com/watch?v=fXB6wftmigU) · 49m 33s

This talk explores strategies for optimizing high-performance applications, particularly in the context of massive multiplayer online games, using Amazon ElastiCache Serverless. It highlights how ElastiCache Serverless provides predictable low latency and automatic scaling, alleviating operational burdens like capacity planning and version upgrades. Key optimization techniques discussed include leveraging persistent connections and pipelining to reduce network overhead, implementing key duplication and splitting to mitigate hotspots, and understanding network latency characteristics across different AWS Availability Zones. The speakers also delve into the architectural design of ElastiCache Serverless, explaining how its proxy layer, vertical scaling, horizontal scaling with warm pools, and IO threads contribute to achieving millions of requests per second with sub-millisecond latencies.

**Technologies:** Amazon ElastiCache Serverless, Valkey, Redis, Node.js, TCP (Transmission Control Protocol)

### 28. AWS re:Invent 2025 - Binge-worthy: Netflix's journey to Amazon Aurora at scale (DAT322)

[https://www.youtube.com/watch?v=cxXG5fZ5wik](https://www.youtube.com/watch?v=cxXG5fZ5wik) · 21m 18s

Netflix's data platform team migrated dozens of self-managed, distributed PostgreSQL-compatible databases to AWS-managed Amazon Aurora PostgreSQL. The migration process prioritized minimizing work for application teams by leveraging AWS Database Migration Service (DMS) and extensive custom tooling for schema analysis, SQL sampling, data validation (batch and live streaming), and automated cutovers. This shift resulted in improved latencies and the proactive identification and resolution of numerous data corruption bugs.

**Technologies:** PostgreSQL, Amazon Aurora PostgreSQL, AWS Database Migration Service (DMS), Apache Kafka, Apache Flink

### 29. AWS re:Invent 2025 - Data modeling core concepts for Amazon DynamoDB (DAT311)

[https://www.youtube.com/watch?v=oMVRvg_8m_I](https://www.youtube.com/watch?v=oMVRvg_8m_I) · 45m 1s

This talk provides a comprehensive overview of data modeling concepts and scaling techniques for Amazon DynamoDB. It begins with an analogy comparing DynamoDB's core concepts (Partition Keys and Sort Keys) to phonebooks and indexes. The session then delves into advanced techniques like Local Secondary Indexes (LSIs) and Global Secondary Indexes (GSIs), explaining their consistency models and query capabilities. Practical scaling strategies are discussed, including client-side compression, splitting large items, soft deletes using DynamoDB Streams and Lambda, and server-side search integration with Amazon OpenSearch Service. The speaker also highlights the importance and cost-effectiveness of Point-in-Time Recovery (PITR) for disaster recovery.

**Technologies:** Amazon DynamoDB, Local Secondary Index (LSI), Global Secondary Index (GSI), Amazon DynamoDB Streams, AWS Lambda

### 30. AWS re:Invent 2025 - Amazon Aurora DSQL: A developer's perspective  (DAT401)

[https://www.youtube.com/watch?v=MNrPJHQaU24](https://www.youtube.com/watch?v=MNrPJHQaU24) · 54m 7s

This talk provides a developer's perspective on Amazon Aurora DSQL, a fast, serverless, and distributed SQL database compatible with PostgreSQL. It highlights DSQL's benefits such as virtually unlimited scale, high availability, and zero infrastructure management. The speakers demonstrate creating a DSQL cluster, connecting via query editor and CloudShell, and building a banking application using AWS Lambda and the AWS CDK. A significant portion of the talk focuses on DSQL's transaction management, optimistic concurrency control (OCC), and strong snapshot isolation, illustrating how to handle concurrent transactions and retries effectively in application code.

**Technologies:** Amazon Aurora DSQL, PostgreSQL, AWS Lambda, AWS Cloud Development Kit (CDK), AWS CloudShell

