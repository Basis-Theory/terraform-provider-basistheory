package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"regexp"
)

func resourceBasisTheoryReactorFormula() *schema.Resource {
	var (
		reactorFormulaTypes       = []string{"private", "official"}
		configurationOptionsTypes = []string{"string", "boolean", "number"}
	)

	return &schema.Resource{
		Description: "Reactor Formula https://docs.basistheory.com/#reactor-formulas",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceReactorFormulaCreate,
		ReadContext:   resourceReactorFormulaRead,
		UpdateContext: resourceReactorFormulaUpdate,
		DeleteContext: resourceReactorFormulaDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Reactor Formula",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the Reactor Formula",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description:  "Private if the Reactor Formula is isolated to a specific tenant. Official if the Reactor Formula is globally available.",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(reactorFormulaTypes, true),
				Required:     true,
			},
			"description": {
				Description: "Description of what the Reactor Formula does",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"icon": {
				Description: "Base64 image format",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"code": {
				Description: "The code that is executed when the Reactor runs",
				Type:        schema.TypeString,
				Required:    true,
			},
			"configuration": {
				Description: "Configuration options required to implement this Reactor Formula",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description:  "Name of the configuration option that will be configured and available in the Reactor context",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Z_]+"), "Configuration name can only contain uppercase letters and '_'"),
						},
						"description": {
							Description: "Description of what the configuration option is for and/or possible values",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},
						"type": {
							Description:  "Data type of the configuration value when configuring a Reactor from this formula",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(configurationOptionsTypes, false),
						},
					},
				},
			},
			"request_parameter": {
				Description: "Request parameters needed at time of reaction",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Name of the request parameter that will be configured and available in the Reactor context",
							Type:        schema.TypeString,
							Required:    true,
						},
						"description": {
							Description: "Description of what the request parameter is for and/or possible values",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
						},
						"type": {
							Description:  "Data type of the request parameter value at the time of reaction",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(configurationOptionsTypes, false),
						},
						"optional": {
							Description: "Whether the request parameter is optional",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
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

func resourceReactorFormulaCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactorFormula := getReactorFormulaFromData(data)

	createReactorFormulaModel := basistheory.CreateReactorFormulaModel{}
	createReactorFormulaModel.SetName(reactorFormula.GetName())
	createReactorFormulaModel.SetDescription(reactorFormula.GetDescription())
	createReactorFormulaModel.SetType(reactorFormula.GetType())
	createReactorFormulaModel.SetCode(reactorFormula.GetCode())
	createReactorFormulaModel.SetIcon(reactorFormula.GetIcon())
	createReactorFormulaModel.SetConfiguration(reactorFormula.GetConfiguration())
	createReactorFormulaModel.SetRequestParameters(reactorFormula.GetRequestParameters())

	createdReactorFormula, _, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulaCreate(ctxWithApiKey).CreateReactorFormulaModel(createReactorFormulaModel).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor Formula:", err)
	}

	data.SetId(createdReactorFormula.GetId())

	return resourceReactorFormulaRead(ctx, data, meta)
}

func resourceReactorFormulaRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactorFormula, _, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulaGetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor Formula:", err)
	}

	data.SetId(reactorFormula.GetId())

	for reactorFormulaDatumName, reactorFormulaDatum := range map[string]interface{}{
		"name":              reactorFormula.GetName(),
		"type":              reactorFormula.GetType(),
		"description":       reactorFormula.GetDescription(),
		"code":              reactorFormula.GetCode(),
		"icon":              reactorFormula.GetIcon(),
		"configuration":     flattenReactorFormulaConfigurationData(reactorFormula.GetConfiguration()),
		"request_parameter": flattenReactorFormulaRequestParameterData(reactorFormula.GetRequestParameters()),
		"created_at":        reactorFormula.GetCreatedAt().String(),
		"created_by":        reactorFormula.GetCreatedBy(),
		"modified_at":       reactorFormula.GetModifiedAt().String(),
		"modified_by":       reactorFormula.GetModifiedBy(),
	} {
		err := data.Set(reactorFormulaDatumName, reactorFormulaDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorFormulaUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	reactorFormula := getReactorFormulaFromData(data)
	updateReactorFormulaModel := basistheory.UpdateReactorFormulaModel{}
	updateReactorFormulaModel.SetName(reactorFormula.GetName())
	updateReactorFormulaModel.SetDescription(reactorFormula.GetDescription())
	updateReactorFormulaModel.SetType(reactorFormula.GetType())
	updateReactorFormulaModel.SetCode(reactorFormula.GetCode())
	updateReactorFormulaModel.SetIcon(reactorFormula.GetIcon())
	updateReactorFormulaModel.SetConfiguration(reactorFormula.GetConfiguration())
	updateReactorFormulaModel.SetRequestParameters(reactorFormula.GetRequestParameters())

	_, _, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulaUpdate(ctxWithApiKey, reactorFormula.GetId()).UpdateReactorFormulaModel(updateReactorFormulaModel).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor Formula:", err)
	}

	return resourceReactorFormulaRead(ctx, data, meta)
}

func resourceReactorFormulaDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx)
	basisTheoryClient := meta.(*basistheory.APIClient)

	_, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulaDelete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error delete Reactor Formula:", err)
	}

	return nil
}

func getReactorFormulaFromData(data *schema.ResourceData) *basistheory.ReactorFormulaModel {
	reactorFormula := &basistheory.ReactorFormulaModel{}
	reactorFormula.SetId(data.Id())
	reactorFormula.SetName(data.Get("name").(string))
	reactorFormula.SetType(data.Get("type").(string))
	reactorFormula.SetDescription(data.Get("description").(string))
	reactorFormula.SetCode(data.Get("code").(string))
	reactorFormula.SetIcon(data.Get("icon").(string))

	var configOptions []basistheory.ReactorFormulaConfigurationModel
	if dataConfig, ok := data.Get("configuration").(*schema.Set); ok {
		for _, dataConfigOption := range dataConfig.List() {
			configMap := dataConfigOption.(map[string]interface{})
			config := basistheory.ReactorFormulaConfigurationModel{}
			config.SetName(configMap["name"].(string))
			config.SetDescription(configMap["description"].(string))
			config.SetType(configMap["type"].(string))
			configOptions = append(configOptions, config)
		}
	}

	reactorFormula.SetConfiguration(configOptions)

	var requestParams []basistheory.ReactorFormulaRequestParameterModel
	if dataRequestParam, ok := data.Get("request_parameter").(*schema.Set); ok {
		for _, requestParam := range dataRequestParam.List() {
			requestParamMap := requestParam.(map[string]interface{})
			requestParam := basistheory.ReactorFormulaRequestParameterModel{}
			requestParam.SetName(requestParamMap["name"].(string))
			requestParam.SetDescription(requestParamMap["description"].(string))
			requestParam.SetType(requestParamMap["type"].(string))
			requestParam.SetOptional(requestParamMap["optional"].(bool))
			requestParams = append(requestParams, requestParam)
		}
	}

	reactorFormula.SetRequestParameters(requestParams)

	return reactorFormula
}

func flattenReactorFormulaConfigurationData(configurationOptions []basistheory.ReactorFormulaConfigurationModel) []interface{} {
	if configurationOptions != nil {
		var configOptions []interface{}

		for _, option := range configurationOptions {
			configOption := make(map[string]interface{})

			configOption["name"] = option.GetName()
			configOption["description"] = option.GetDescription()
			configOption["type"] = option.GetType()

			configOptions = append(configOptions, configOption)
		}

		return configOptions
	}

	return make([]interface{}, 0)
}

func flattenReactorFormulaRequestParameterData(requestParameters []basistheory.ReactorFormulaRequestParameterModel) []interface{} {
	if requestParameters != nil {
		var parameterData []interface{}

		for _, param := range requestParameters {
			parameterDatum := make(map[string]interface{})

			parameterDatum["name"] = param.GetName()
			parameterDatum["description"] = param.GetDescription()
			parameterDatum["type"] = param.GetType()
			parameterDatum["optional"] = param.GetOptional()

			parameterData = append(parameterData, parameterDatum)
		}

		return parameterData
	}

	return make([]interface{}, 0)
}
