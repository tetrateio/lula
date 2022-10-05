# NIST Demo

Just to re-iterate what has been mentioned:
- (Navigate to the repo1 Big Bang group)
    - We have this concept on the Big Bang platform - It's a collection of tool packages that provide a baseline implementation for the DoD DevSecOps reference design
    - Any one tool can bring with it controls that it satisfies that the environment can leverage
    - A package provides a tool-relevant universe of controls that CAN be satisfied to some extent by the tool itself
        - I'd call this the first pillar of three for compliance as code - it creates an machine-readable document for inheritance by the consumers - which is version-controlled.
    - These component definitions can be aggregated such that the platform can provide a subset of controls immediately inheritable
        - This doesn't mean they are inheritable - just that they can be if properly configured. 

- (Navigate to RegScale)
    - Moving over to visualization - We've partnered with RegScale on this presentation to really show the power of OSCAL  
    - We've then taken that machine-readable OSCAL component definition and imported it into GRC tooling for visualization
        - Here i'd call visualization as the third pillar (bear with me on why I skipped the 2nd)
            - Visualization allows multiple personas to display large amounts of data in ways that cna be easily traced between one-another
            - It also provides a consolidated view
    - Tooling such as RegScale (which is optimized around OSCAL) can play a huge role in platform releases and tracing control updates.
    - API driven tooling will support automation in many ways
        - Uploading the static artifacts (version controlled OSCAL component definitions)
        - Supporting 3rd party tools to integrate for automation

- So what is the 2nd pillar that divides Version-controlled OSCAL and the Visualization?
    - It's actual validation - reciprocity is, for lack of a better word, silly (I can go into more details on that later)
    - We need insight into what controls are satisfied, given the actual configuration and environment state. 

- What I will be demonstrating today is one solution for that second pillar
    - That is - how does my live environment reflect the controls that have been said to be satisfied?
    - Let me preface the demonstration  with a few points:
        - We're here to attempt to articulate this second pillar and the importance that OSCAL plays in solving the broader accredition and compliance domain.
        - We're demonstrating the capability, given a proof-of-concept tool, to try and outline the problem and generate community interest
            - There is still work to be done on the OSCAL documents and establishing compliance with the schemas

- (Navigate to the Local project directory)
    - If compliance can be validated with automation - you can essentially audit your system(s) continuously
    - (Show oscal-component.yaml)
        - What I have locally is a modified version of the Component definition for Istio - the platform service-mesh tool
            - It includes a field that automation can ingest and validate against a live kubernetes
            - The Istio component definition specifies in AC-4 that:
                - "The information system enforces approved authorizations for controlling the flow of information within the system and between interconnected systems"
            - As a lightweight method to mock testing that this control is satisfied, we are going to validate that all pods are istio-injected and that istio is configured to only allow MTLS traffic between pods.
    
    - (Open Terminal)
        - Now for my live environment - I have the big bang platform deployed with a number of workloads running
            - Logging stacks
            - Monitoring stacks
            - runtime security
            - policy enforcement
            - CI/CD tools
            - etc
        - If we run the tool, we will see an short output - followed by it uploading the artifacts to RegScale
        - `./compliance-auditor execute test/cli/component-definitions/oscal-component.yaml`
        - Here we can see a quick pass/fail - but lets see how RegScale has connected this assessment to the control and component

        - (Switch to RegScale)
            - From the component, we can navigate through the scorecard - noting things like overall compliance
            - We can then dial into our control and assessments where we will see the current state of the latest assessment.
                - Viewing one of these assessments reveals how RegScale documents how the valdiation is performed, all via submission through API's

        - (Switch back to terminal)
            - Now we'll purposefully inject a resource into our environment that violates the global configuration required for satisfying the control
            - `kubectl run test --image=nginx:1.21.3`
            - We'll see that this workload enters a running state `kubectl get po -n default`
            - We'll now re-run the tool
                - `./compliance-auditor execute test/cli/component-definitions/oscal-component.yaml`
                - Note the fail status

        - (Switch back to RegScale)
            - We'll again navigate through the component, scorecard, control and see that it annotates a now failing state. 


        - Many plans for where we hope this tool can evolve and be valuable to the greater community. 
            - Very easily it can be scheduled to audit periodically
        
        - Questions??

    
