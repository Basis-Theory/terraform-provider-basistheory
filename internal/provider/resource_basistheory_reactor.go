package provider

import (
	"context"
	basistheory "github.com/Basis-Theory/go-sdk/v3"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v3/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryReactor() *schema.Resource {
	return &schema.Resource{
		Description: "Reactor https://docs.basistheory.com/docs/api/reactors",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceReactorCreate,
		ReadContext:   resourceReactorRead,
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
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor := getReactorFromData(data)

	createReactorRequest := &basistheory.CreateReactorRequest{
		Name:          getStringValue(reactor.Name),
		Code:          getStringValue(reactor.Code),
		Configuration: reactor.Configuration,
		Application:   reactor.Application,
	}

	createdReactor, err := basisTheoryClient.Reactors.Create(ctx, createReactorRequest)

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor:", err)
	}

	data.SetId(*createdReactor.ID)

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor, err := basisTheoryClient.Reactors.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor:", err)
	}

	data.SetId(*reactor.ID)

	application := reactor.Application

	modifiedAt := ""

	if reactor.ModifiedAt != nil {
		modifiedAt = reactor.ModifiedAt.String()
	}

	for reactorDatumName, reactorDatum := range map[string]interface{}{
		"tenant_id": reactor.TenantID,
		"name":      reactor.Name,
		"code":      reactor.Code,
		"application_id": func() interface{} {
			if application != nil {
				return application.ID
			}
			return nil
		}(),
		"configuration": reactor.Configuration,
		"created_at":    reactor.CreatedAt.String(),
		"created_by":    reactor.CreatedBy,
		"modified_at":   modifiedAt,
		"modified_by":   reactor.ModifiedBy,
	} {
		err := data.Set(reactorDatumName, reactorDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor := getReactorFromData(data)
	updateReactorRequest := &basistheory.UpdateReactorRequest{
		Name:          getStringValue(reactor.Name),
		Code:          getStringValue(reactor.Code),
		Configuration: reactor.Configuration,
		Application:   reactor.Application,
	}
	_, err := basisTheoryClient.Reactors.Update(ctx, *reactor.ID, updateReactorRequest)

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor:", err)
	}

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := basisTheoryClient.Reactors.Delete(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error deleting Reactor:", err)
	}

	return nil
}

func getReactorFromData(data *schema.ResourceData) *basistheory.Reactor {
	reactor := &basistheory.Reactor{}
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

	application := &basistheory.Application{}
	applicationId := data.Get("application_id").(string)
	if applicationId != "" {
		application.ID = getStringPointer(applicationId)
		reactor.Application = application
	}

	return reactor
}
