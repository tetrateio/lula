# 5. cel-provider

**Date:** 2023-10-03

## Status

Proposed

## Context

The [Common Expression Language](https://github.com/google/cel-spec) (CEL) is a domain-specific language used for evaluating expressions or conditions in various software applications and systems. It is designed to provide a consistent and standardized way to perform evaluations, filtering, and decision-making within these systems. CEL is often used in contexts such as policy enforcement, filtering data, or defining rules in a concise and easy-to-understand manner. It's particularly valuable in applications where dynamic, user-defined expressions or conditions are needed to determine outcomes or control behaviors.

## Decision

The decision is to integrate CEL as a provider within Lula through the use of Go CEL libraries. This enhancement will bring tremendous value to Lula by:

- **Empowering Lula operators:** By integrating CEL, we empower Lula operators with advanced capabilities to write validation policies that align precisely with operator skills.

- **Flexibility:** Lula operators will have flexibility in defining and utilizing validation measures. These measures can be applied either in isolation or seamlessly integrated with other validating providers. This means Lula will be adaptable to a wider array of scenarios and use cases, ensuring it meets the diverse needs of our users.

## Consequences

The integration of CEL as a provider within Lula will result in the following substantial benefits:

- **Expanded Language Options:** Lula operators will now have access to a wider range of expressive language options for creating validation policies. This means they can choose the language that best suits their specific needs and expertise.

- **Enhanced Decision-Making:** With CEL, Lula users can make more informed decisions, as CEL's powerful evaluation capabilities enable them to create sophisticated and context-aware rules. This leads to improved accuracy and efficiency in policy enforcement.

- **Seamless Integration:** CEL's integration will seamlessly coexist with existing validation providers, allowing Lula to take full advantage of both CEL's capabilities and the strengths of other providers. This interoperability ensures a smooth transition and maximum utility for all users.

- **Adaptability:** Lula becomes more adaptable to various validation requirements, enabling users to configure and fine-tune their validation processes according to evolving needs and industry standards.

The integration of CEL as a provider is a substantial step towards making Lula an even more versatile and powerful tool for our users, unlocking new possibilities and efficiencies in their operations.
