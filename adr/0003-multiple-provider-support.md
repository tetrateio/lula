# 3. multiple-provider-support

Date: 2023-09-26

## Status

Proposed

## Context

In an effort to produce an initial baseline for the project, It is important to establish standards for policy provider integration as it pertains to data requiring collection and how it will be validated. The key aspects covered in this context are:

- **Decision Objective:** The decision's primary goal is to establish standards for integrating policy providers into the codebase while considering the potential need to support multiple validation providers in the future.

- **Code Structure:** Emphasizing the importance of defining a flexible and extensible code structure that can accommodate various validation providers. This structure should encompass hooks, or points of connection, between the command-line interface (CLI) code and the components responsible for data collection and validation.

- **Decoupling:** Decoupling is still a central theme in this context. The idea is to separate the data collection and validation logic from the CLI code.

## Decision

Restructuring the current codebase to support decoupling data collection logic (when required) and validating logic (policy provider) from the CLI code. The decision includes the following key points:

- **Kyverno Validating Logic Removal:** The decision involves the removal of the current Kyverno Validating logic from the codebase, aligning with the aim to support multiple validation providers. This requires Kyverno logic to be re-integrated in the future following established processes and testing.

- **Framework for Integration:** The decision also entails the creation of a flexible and extensible framework that allows for the inclusion of new logic related to data collection and validators. This framework should not be tied to a single validation provider but should support the integration of multiple providers.

- **ADR Requirement:** An important aspect of this decision is the establishment of an Architectural Decision Record (ADR) requirement for any new data collection and validator logic. This documentation should account for the potential compatibility with various validation providers, promoting adaptability.

## Consequences

This section outlines the expected outcomes and impacts of the decision made, taking into account the support for multiple validation providers:

- **Project Flexibility:** The primary consequence emphasized here is the enhanced flexibility of the project. By reorganizing the codebase and ensuring that it accommodates various validation providers, the project becomes versatile and adaptable to future changes or enhancements.

- **Modularity and Extensibility:** The decision will result in a more modular and extensible codebase. This modularity enables easier testing of individual components and the extensibility facilitates the integration of multiple validation providers.

- **Backwards Compatibility:** It's noted that previous installations of the project may not be backwards compatible with the changes introduced by this decision. Existing users may need to update their configurations or processes to align with the new expected structure.

- **Testing and Documentation:** Any integrated logic, such as data collection and validators, will require thorough testing and documentation through an ADR. This documentation should consider compatibility and integration with various validation providers.
