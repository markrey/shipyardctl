package cmd

var TARGET_ENDPOINT = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<TargetEndpoint name="default">
    <Description/>
    <FaultRules/>
    <PreFlow name="PreFlow">
        <Request>
            <Step>
                <Name>RetainHostHeader</Name>
            </Step>
            <Step>
                <Name>SetRoutingAPIKey</Name>
            </Step>
        </Request>
        <Response>
            <Step>
                <Name>AddCors</Name>
            </Step>
        </Response>
    </PreFlow>
    <PostFlow name="PostFlow">
        <Request/>
        <Response/>
    </PostFlow>
    <Flows/>
    <HTTPTargetConnection>
        <Properties/>
        <URL>https://shipyard-backend-west.e2e.apigee.net</URL>
    </HTTPTargetConnection>
</TargetEndpoint>`

var PROXY_ENDPOINT = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<ProxyEndpoint name="default">
    <Description/>
    <FaultRules/>
    <PreFlow name="PreFlow">
        <Request/>
        <Response/>
    </PreFlow>
    <PostFlow name="PostFlow">
        <Request/>
        <Response/>
    </PostFlow>
    <Flows/>
    <HTTPProxyConnection>
        <BasePath>/</BasePath>
        <Properties/>
        <VirtualHost>default</VirtualHost>
        <VirtualHost>secure</VirtualHost>
    </HTTPProxyConnection>
    <RouteRule name="default">
        <TargetEndpoint>default</TargetEndpoint>
    </RouteRule>
</ProxyEndpoint>`

var ADD_CORS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<AssignMessage async="false" continueOnError="false" enabled="true" name="AddCors">
    <DisplayName>Add CORS Headers</DisplayName>
    <FaultRules/>
    <Properties/>
    <Add>
        <Headers>
            <Header name="Access-Control-Allow-Origin">*</Header>
            <Header name="Access-Control-Allow-Headers">origin, x-requested-with, accept</Header>
            <Header name="Access-Control-Max-Age">3628800</Header>
            <Header name="Access-Control-Allow-Methods">GET, PUT, POST, DELETE</Header>
        </Headers>
    </Add>
    <IgnoreUnresolvedVariables>true</IgnoreUnresolvedVariables>
    <AssignTo createNew="false" transport="http" type="response"/>
</AssignMessage>`

var RETAIN_HOST = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<AssignMessage async="false" continueOnError="false" enabled="true" name="RetainHostHeader">
    <DisplayName>Retain Host Header for Target</DisplayName>
    <Properties/>
    <AssignVariable>
        <Name>target.header.host</Name>
        <Value>{{.Org}}-{{.Env}}.apigee.net</Value>
        <Ref/>
    </AssignVariable>
    <IgnoreUnresolvedVariables>true</IgnoreUnresolvedVariables>
    <AssignTo createNew="false" transport="http" type="request"/>
</AssignMessage>`

var ROUTING_KEY = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<AssignMessage async="false" continueOnError="false" enabled="true" name="SetRoutingAPIKey">
    <DisplayName>Set Routing API Key</DisplayName>
    <Properties/>
    <Set>
        <Headers>
            <Header name="X-ROUTING-API-KEY">{{.PublicKey}}</Header>
        </Headers>
    </Set>
    <IgnoreUnresolvedVariables>true</IgnoreUnresolvedVariables>
    <AssignTo createNew="false" transport="http" type="request"/>
</AssignMessage>`

var PROXY_XML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<APIProxy revision="1" name="{{.AppName}}">
    <ConfigurationVersion majorVersion="4" minorVersion="0"/>
    <CreatedAt>1459886430613</CreatedAt>
    <CreatedBy>shipyard@apigee.com</CreatedBy>
    <Description>This is a proxy for {{.AppName}} deployed on Shipyard.</Description>
    <DisplayName>{{.AppName}}</DisplayName>
    <LastModifiedAt>1459886430613</LastModifiedAt>
    <LastModifiedBy>shipyard@apigee.com</LastModifiedBy>
    <Policies>
        <Policy>AddCors</Policy>
        <Policy>RetainHostHeader</Policy>
        <Policy>SetRoutingAPIKey</Policy>
    </Policies>
    <ProxyEndpoints>
        <ProxyEndpoint>default</ProxyEndpoint>
    </ProxyEndpoints>
    <Resources/>
    <TargetServers/>
    <TargetEndpoints>
        <TargetEndpoint>default</TargetEndpoint>
    </TargetEndpoints>
    <validate>false</validate>
</APIProxy>
`