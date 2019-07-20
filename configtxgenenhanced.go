package main

const _configTxTemplateV13 = `
Organizations:
    - &OrdererOrg
            Name: {{index .orderers "mspID" }}
            ID: {{index .orderers "mspID" }}
            MSPDir: crypto-config/ordererOrganizations/{{ index .orderers "domain" }}/msp
            Policies:
                Readers:
                    Type: Signature
                    Rule: "OR('{{index .orderers "mspID" }}.member')"
                Writers:
                    Type: Signature
                    Rule: "OR('{{index .orderers "mspID" }}.member')"
                Admins:
                    Type: Signature
                    Rule: "OR('{{index .orderers "mspID" }}.admin')"

    {{range .orgs}}
    - &{{ .name}}Org
            Name: {{.mspID}}
            ID: {{.mspID}}
            MSPDir: crypto-config/peerOrganizations/{{ .domain  }}/msp
            Policies:
                Readers:
                    Type: Signature
                    Rule: "OR('{{.mspID}}.admin', '{{.mspID}}.peer', '{{.mspID}}.client' )"
                Writers:
                    Type: Signature
                    Rule: "OR('{{.mspID}}.admin', '{{.mspID}}.client' )"
                Admins:
                    Type: Signature
                    Rule: "OR('{{.mspID}}.admin')"
            AnchorPeers:
              - Host: peer0.{{.domain}}
                Port: 7051
    {{end }}

Capabilities:
    Channel: &ChannelCapabilities
        V1_3: true
    Orderer: &OrdererCapabilities
        V1_1: true
    Application: &ApplicationCapabilities
        V1_3: true
        V1_2: false
        V1_1: false

Application: &ApplicationDefaults
    Organizations:

    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"

    Capabilities:
        <<: *ApplicationCapabilities
{{ if  and (eq .orderers.type "kafka")  (  .orderers.haCount ) }}
Orderer: &OrdererDefaults
    OrdererType: kafka
    Addresses:{{ range .ordererFDQNList }}
          - {{.}}{{end}}
    BatchTimeout: 2s
    BatchSize:
        MaxMessageCount: 10
        AbsoluteMaxBytes: 98 MB
        PreferredMaxBytes: 1024 KB
    Kafka:
        Brokers:
            - kafka0:9092
            - kafka1:9092
            - kafka2:9092
            - kafka3:9092
    Organizations:

    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"
        BlockValidation:
            Type: ImplicitMeta
            Rule: "ANY Writers"
    Capabilities:
        <<: *OrdererCapabilities
{{else}}
Orderer: &OrdererDefaults
    OrdererType: solo
    Addresses:
          - {{index .orderers "ordererHostname" }}.{{index .orderers "domain"}}:7050
    BatchTimeout: 2s
    BatchSize:
        MaxMessageCount: 16
        AbsoluteMaxBytes: 98 MB
        PreferredMaxBytes: 1024 KB
    Kafka:
        Brokers:
            - 127.0.0.1:9092
    Organizations:

    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"
        BlockValidation:
            Type: ImplicitMeta
            Rule: "ANY Writers"
    Capabilities:
        <<: *OrdererCapabilities
{{end}}
Channel: &ChannelDefaults
    Policies:
        Readers:
            Type: ImplicitMeta
            Rule: "ANY Readers"
        Writers:
            Type: ImplicitMeta
            Rule: "ANY Writers"
        Admins:
            Type: ImplicitMeta
            Rule: "MAJORITY Admins"
    Capabilities:
        <<: *ChannelCapabilities

Profiles:
    OrdererGenesis:
        <<: *ChannelDefaults
        Orderer:
            <<: *OrdererDefaults
            Organizations:
                - *OrdererOrg
            Capabilities:
                <<: *OrdererCapabilities 
        Consortiums:
            {{.consortium}}:
                Organizations:
                   {{ range .orgs}}- *{{ .name}}Org
                   {{end}}
        {{ $x :=.consortium}}
    {{range .channels}}
    {{.channelName}}:
        Consortium: {{$x}}
        Application:
            <<: *ApplicationDefaults
            Organizations:
              {{range $index,$var := .orgs}}- *{{$var}}Org
              {{end}}
            Capabilities:
              <<: *ApplicationCapabilities
    {{end}}
        

`
