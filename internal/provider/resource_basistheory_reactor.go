package provider

import (
	"context"
	basistheoryV2 "github.com/Basis-Theory/go-sdk"

	"github.com/Basis-Theory/basistheory-go/v6"
	basistheoryV2client "github.com/Basis-Theory/go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryReactor() *schema.Resource {
	return &schema.Resource{
		Description: "Reactor https://docs.basistheory.com/docs/api/reactors",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceReactorCreateV2,
		ReadContext:   resourceReactorReadV2,
		UpdateContext: resourceReactorUpdate,
		DeleteContext: resourceReactorDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the Reactor",
				Type:        schema.TypeString,
				Required:    true,
			},
			"code": {
				Description: "The code that is executed when the Reactor runs",
				Type:        schema.TypeString,
				Required:    true,
			},
			"application_id": {
				Description: "The Application's permissions used in the BasisTheory instance passed into the Reactor",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"tenant_id": {
				Description: "Tenant identifier where this Reactor was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"configuration": {
				Description: "Configuration for the Reactor",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created_at": {
				Description: "Timestamp at which the Reactor was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_at": {
				Description: "Timestamp at which the Reactor was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_by": {
				Description: "Identifier for who last modified the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceReactorCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactor := getReactorFromData(data)

	createReactorRequest := *basistheory.NewCreateReactorRequest(reactor.GetName(), reactor.GetCode())
	createReactorRequest.SetConfiguration(reactor.GetConfiguration())

	if application, ok := reactor.GetApplicationOk(); ok {
		createReactorRequest.SetApplication(*application)
	}

	createdReactor, response, err := basisTheoryClient.ReactorsApi.Create(ctxWithApiKey).CreateReactorRequest(createReactorRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor:", response, err)
	}

	data.SetId(createdReactor.GetId())

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorCreateV2(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	reactor := getReactorFromDataV2(data)

	createReactorRequest := &basistheoryV2.CreateReactorRequest{
		//reactor.GetName(), reactor.GetCode()
		Name: getStringValue(reactor.Name),
		Code: getStringValue(reactor.Code),
	}
	createReactorRequest.Configuration = reactor.Configuration
	createReactorRequest.Application = reactor.Application

	createdReactor, err := basisTheoryClient.Reactors.Create(ctx, createReactorRequest)

	if err != nil {
		return apiErrorDiagnosticsV2("Error creating Reactor:", err)
	}

	data.SetId(*createdReactor.ID)

	return resourceReactorReadV2(ctx, data, meta)
}

func resourceReactorRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactor, response, err := basisTheoryClient.ReactorsApi.GetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor:", response, err)
	}

	data.SetId(reactor.GetId())

	application := reactor.GetApplication()

	modifiedAt := ""

	if reactor.ModifiedAt.IsSet() {
		modifiedAt = reactor.GetModifiedAt().String()
	}

	for reactorDatumName, reactorDatum := range map[string]interface{}{
		"tenant_id":      reactor.GetTenantId(),
		"name":           reactor.GetName(),
		"code":           reactor.GetCode(),
		"application_id": application.GetId(),
		"configuration":  reactor.GetConfiguration(),
		"created_at":     reactor.GetCreatedAt().String(),
		"created_by":     reactor.GetCreatedBy(),
		"modified_at":    modifiedAt,
		"modified_by":    reactor.GetModifiedBy(),
	} {
		err := data.Set(reactorDatumName, reactorDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorReadV2(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	reactor, err := basisTheoryClient.Reactors.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnosticsV2("Error reading Reactor:", err)
	}

	data.SetId(*reactor.ID)

	application := reactor.Application

	modifiedAt := ""

	if reactor.ModifiedAt != nil {
		modifiedAt = reactor.ModifiedAt.String()
	}

	for reactorDatumName, reactorDatum := range map[string]interface{}{
		"tenant_id":      reactor.TenantID,
		"name":           reactor.Name,
		"code":           reactor.Code,
		"application_id": func() interface{} {
			if application != nil {
				return application.ID
			}
			return nil
		}(),
		"configuration":  reactor.Configuration,
		"created_at":     reactor.CreatedAt.String(),
		"created_by":     reactor.CreatedBy,
		"modified_at":    modifiedAt,
		"modified_by":    reactor.ModifiedBy,
	} {
		err := data.Set(reactorDatumName, reactorDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactor := getReactorFromData(data)
	updateReactorRequest := *basistheory.NewUpdateReactorRequest(reactor.GetName(), reactor.GetCode())
	updateReactorRequest.SetConfiguration(reactor.GetConfiguration())

	if application, ok := reactor.GetApplicationOk(); ok {
		updateReactorRequest.SetApplication(*application)
	}

	_, response, err := basisTheoryClient.ReactorsApi.Update(ctxWithApiKey, reactor.GetId()).UpdateReactorRequest(updateReactorRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor:", response, err)
	}

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ReactorsApi.Delete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error deleting Reactor:", response, err)
	}

	return nil
}

func getReactorFromData(data *schema.ResourceData) *basistheory.Reactor {
	reactor := basistheory.NewReactor()
	reactor.SetId(data.Id())
	reactor.SetName(data.Get("name").(string))

	reactorCode := data.Get("code").(string)
	if reactorCode != "" {
		reactor.SetCode(reactorCode)
	}

	configOptions := map[string]string{}
	for key, value := range data.Get("configuration").(map[string]interface{}) {
		configOptions[key] = value.(string)
	}

	reactor.SetConfiguration(configOptions)

	application := *basistheory.NewApplication()
	applicationId := data.Get("application_id").(string)
	if applicationId != "" {
		application.SetId(applicationId)
		reactor.SetApplication(application)
	}

	return reactor
}

func getReactorFromDataV2(data *schema.ResourceData) *basistheoryV2.Reactor {
	reactor := &basistheoryV2.Reactor{}
	reactor.ID = getStringPointer(data.Id())
	reactor.Name = getStringPointer(data.Get("name"))

	reactorCode := data.Get("code").(string)
	if reactorCode != "" {
		reactor.Code = getStringPointer(reactorCode)
	}

	configOptions := map[string]*string{}
	for key, value := range data.Get("configuration").(map[string]interface{}) {
		configOptions[key] = getStringPointer(value)
	}

	reactor.Configuration = configOptions

	application := &basistheoryV2.Application{}
	applicationId := data.Get("application_id").(string)
	if applicationId != "" {
		application.ID = getStringPointer(applicationId)
		reactor.Application = application
	}

	return reactor
}
