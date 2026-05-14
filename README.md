# Doctryne contributions analysis tool

## Rules

List of supported rules:

- [ ] mail domains
- [ ] mail domain countries
- [ ] last version usage
- [ ] license type
- [ ] contributors' counties and domains
- [ ] toxic contributors
- [ ] toxic phrases in description/readme or code
- [ ] documentation language
- [ ] source code repository
- [ ] source code comments language
- [ ] registry links, domains, email
- [ ] github.com stars
- [ ] project age
- [ ] project last update
- [ ] project archivation status
- [ ] project funding status
- [ ] project category/labels/badges

## Data extraction pipelines

- JavaScript/TypeScript
    - package-lock.json
        - registry.npmjs.org
            - id
            - description
            - keywords
            - homepage
            - official author
                - name
                - email
            - official contributor/maintainers information
                - names
                - emails
            - git repository
                - github.com info
            - revisions/versions
            - timestamps
                - publication date
                - last modification date
- Golang
    - go.mod
        - github.com info
        - pkg.go.dev info
            - timestamps
                - publication date
                - last modification date
            - imports
            - imported by
- Java
    - pom.xml
        - mvnrepository.com
            - license
            - categories
            - homepage
            - github.com info
            - imports
            - imported by

### GitHub data extraction (github.com)

- stars
- forks
- license
- readme
- archive status
- sponsors
- languages
- homepage
- actions
    - workflows
- issues
    - open issues
    - lately closed issues
- security
    - max vulnerability cvss
    - total number of open vulnerabilities
    - oldest open vulnerability
- pull requests
    - open pull requests
    - lately closed pull requests
- activity
    - first commit
    - latest commit
    - number of commits
- contributors
    - name
    - email
    - country/region
    - other projects
    - other companies
- 