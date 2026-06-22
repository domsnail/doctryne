# Application components

1. [InspectionService](../internal/service/inspect_service/service.go)
2. [ManifestService](../internal/service/manifest_service/service.go)
3. [RegistryServices](../internal/service/registry_service/service.go)
    - NPM
    - Golang
4. [GitHubService](../internal/service/github_service/service.go)

```mermaid
---
title: Data flow
config:
    htmlLabels: false
---
sequenceDiagram
   User Input ->> Inspection Service: Scan options
   Note right of User Input: Through CLI, REST or gRPC API
   activate Inspection Service
   Inspection Service ->> Inspection Service: Search for manifests
   Inspection Service ->> Manifest Service: Manifests for analysis
   deactivate Inspection Service
   activate Manifest Service
   Manifest Service ->> Inspection Service: Packages for each manifests
   deactivate Manifest Service
%%    
   activate Inspection Service
   Inspection Service ->> Inspection Service: Analyse duplications, prepare payloads
   Inspection Service ->> Registry Services: Search packages
%% 
   activate Registry Services
   Note right of Registry Service: Registry clients use intercepting cache
   Registry Services -->> NPM/Gopkg/Maven: Fetch package metadata
   NPM/Gopkg/Maven -->> Registry Services: Package metadata (contributors, descriptions etc.)
   Registry Services -->> NPM/Gopkg/Maven: Fetch package stats
   NPM/Gopkg/Maven -->> Registry Services: Package stats by period
   Registry Services ->> Inspection Service: Packages info, contributors, stats
   deactivate Registry Services
%% 
   Note right of Inspection Service: Building profiles for each Developer
   Inspection Service ->> Inspection Service: Analyse duplications, prepare payloads
   Inspection Service ->> Github Service: Search developers, repositories
   activate Github Service
   Github Service ->> Inspection Service: Developers info, repository data, commits, stats
   deactivate Github Service
   Inspection Service ->> Inspection Service: Compacting, aggregating, clearing data
   Note right of Inspection Service: Rulesets and provided options are used
   Inspection Service ->> Inspection Service: Evaluating collected data, scoring all packages
   Inspection Service ->> Inspection Service: Building report
   deactivate Inspection Service
%%    
   Inspection Service ->> User Input: Evaluations report
```

