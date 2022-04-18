package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryReactor() *schema.Resource {
	return &schema.Resource{
		Description: "Reactor https://docs.basistheory.com/#reactors",

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
			"formula_id": {
				Description: "Reactor Formula for the Reactor",
				Type:        schema.TypeString,
				Required:    true,
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
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactor := getReactorFromData(data)

	createReactorModel := basistheory.CreateReactorModel{}
	createReactorModel.SetName(reactor.GetName())
	createReactorModel.SetFormula(reactor.GetFormula())
	createReactorModel.SetConfiguration(reactor.GetConfiguration())

	createdReactor, _, err := basisTheoryClient.ReactorsApi.ReactorCreate(ctxWithApiKey).CreateReactorModel(createReactorModel).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor:", err)
	}

	data.SetId(createdReactor.GetId())

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactor, _, err := basisTheoryClient.ReactorsApi.ReactorGetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor:", err)
	}

	data.SetId(reactor.GetId())

	reactorFormula := reactor.GetFormula()

	for reactorDatumName, reactorDatum := range map[string]interface{}{
		"tenant_id":     reactor.GetTenantId(),
		"name":          reactor.GetName(),
		"formula_id":    reactorFormula.GetId(),
		"configuration": reactor.GetConfiguration(),
		"created_at":    reactor.GetCreatedAt().String(),
		"created_by":    reactor.GetCreatedBy(),
		"modified_at":   reactor.GetModifiedAt().String(),
		"modified_by":   reactor.GetModifiedBy(),
	} {
		err := data.Set(reactorDatumName, reactorDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactor := getReactorFromData(data)
	updateReactorModel := basistheory.UpdateReactorModel{}
	updateReactorModel.SetName(reactor.GetName())
	updateReactorModel.SetConfiguration(reactor.GetConfiguration())

	_, _, err := basisTheoryClient.ReactorsApi.ReactorUpdate(ctxWithApiKey, reactor.GetId()).UpdateReactorModel(updateReactorModel).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor:", err)
	}

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	_, err := basisTheoryClient.ReactorsApi.ReactorDelete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error deleting Reactor:", err)
	}

	return nil
}

func getReactorFromData(data *schema.ResourceData) *basistheory.ReactorModel {
	reactor := &basistheory.ReactorModel{}
	reactor.SetId(data.Id())
	reactor.SetName(data.Get("name").(string))

	reactorFormula := basistheory.ReactorFormulaModel{}
	reactorFormula.SetId(data.Get("formula_id").(string))

	reactor.SetFormula(reactorFormula)

	configOptions := map[string]string{}
	for key, value := range data.Get("configuration").(map[string]interface{}) {
		configOptions[key] = value.(string)
	}

	reactor.SetConfiguration(configOptions)

	return reactor
}
