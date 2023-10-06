# 2. OSCAL as default data format

Date: 2023-04-24

## Status

Accepted

## Context

Our project requires a standardized data format for both input and output that can be easily consumed by existing or future tools in the risk management and compliance ecosystem. The amount of time that is required by security expertise in the domain of capturing, assessing, and validating control information is significant; as such, the data format should be standardized and based upon best practices for both documentation as well as posture towards automation. The standard should optimally be maintained by an authority in this domain and this project will aim to augment the adoption of said standard/data-format. 

## Decision

We decided to use NIST's Open Security Controls Assessment Language ([OSCAL](https://pages.nist.gov/OSCAL/)) as the chosen default data format for this project. 

### Reasons for using OSCAL
1. Standardized format for Control-Based Risk Management support
2. Maintained by NIST
3. Machine-readable format
4. Allows the project to foster OSCAL adoption for end-users
5. Re-usable and Open Source artifacts can be contributed to upstream projects
6. Format is accepted by a number of GRC tools to-date 

## Consequences

- Golang datatype support is limited for OSCAL. This will require creation and maintenance.
- The project will not need to prescribe or maintain a custom data-format and can augment the adoption of OSCAL.
- Integration with the project outputs can be based upon upstream data models.
- Our developers will need to track updates to OSCAL going forward. 
