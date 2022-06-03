package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go/v2"
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
				Description: "The code that is executed when the Reactor runs. Set to empty string to indicate Reactor Formula as `coming_soon`",
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
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactorFormula := getReactorFormulaFromData(data)

	createReactorFormulaRequest := *basistheory.NewCreateReactorFormulaRequest(reactorFormula.GetType(), reactorFormula.GetName())
	createReactorFormulaRequest.SetDescription(reactorFormula.GetDescription())
	createReactorFormulaRequest.SetCode(reactorFormula.GetCode())
	createReactorFormulaRequest.SetIcon(reactorFormula.GetIcon())
	createReactorFormulaRequest.SetConfiguration(reactorFormula.GetConfiguration())
	createReactorFormulaRequest.SetRequestParameters(reactorFormula.GetRequestParameters())

	createdReactorFormula, response, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulasCreate(ctxWithApiKey).CreateReactorFormulaRequest(createReactorFormulaRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor Formula:", response, err)
	}

	data.SetId(createdReactorFormula.GetId())

	return resourceReactorFormulaRead(ctx, data, meta)
}

func resourceReactorFormulaRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactorFormula, response, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulasGetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor Formula:", response, err)
	}

	data.SetId(reactorFormula.GetId())

	modifiedAt := ""

	if reactorFormula.ModifiedAt.IsSet() {
		modifiedAt = reactorFormula.GetModifiedAt().String()
	}

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
		"modified_at":       modifiedAt,
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
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	reactorFormula := getReactorFormulaFromData(data)
	updateReactorFormulaRequest := *basistheory.NewUpdateReactorFormulaRequest(reactorFormula.GetType(), reactorFormula.GetName())
	updateReactorFormulaRequest.SetDescription(reactorFormula.GetDescription())
	updateReactorFormulaRequest.SetCode(reactorFormula.GetCode())
	updateReactorFormulaRequest.SetIcon(reactorFormula.GetIcon())
	updateReactorFormulaRequest.SetConfiguration(reactorFormula.GetConfiguration())
	updateReactorFormulaRequest.SetRequestParameters(reactorFormula.GetRequestParameters())

	_, response, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulasUpdate(ctxWithApiKey, reactorFormula.GetId()).UpdateReactorFormulaRequest(updateReactorFormulaRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor Formula:", response, err)
	}

	return resourceReactorFormulaRead(ctx, data, meta)
}

func resourceReactorFormulaDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulasDelete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error delete Reactor Formula:", response, err)
	}

	return nil
}

func getReactorFormulaFromData(data *schema.ResourceData) *basistheory.ReactorFormula {
	reactorFormula := basistheory.NewReactorFormula()
	reactorFormula.SetId(data.Id())
	reactorFormula.SetName(data.Get("name").(string))
	reactorFormula.SetType(data.Get("type").(string))
	reactorFormula.SetDescription(data.Get("description").(string))
	reactorFormula.SetCode(data.Get("code").(string))
	reactorFormula.SetIcon(data.Get("icon").(string))

	var configOptions []basistheory.ReactorFormulaConfiguration
	if dataConfig, ok := data.Get("configuration").(*schema.Set); ok {
		for _, dataConfigOption := range dataConfig.List() {
			configMap := dataConfigOption.(map[string]interface{})
			config := *basistheory.NewReactorFormulaConfiguration(configMap["name"].(string), configMap["type"].(string))
			config.SetDescription(configMap["description"].(string))
			configOptions = append(configOptions, config)
		}
	}

	reactorFormula.SetConfiguration(configOptions)

	var requestParams []basistheory.ReactorFormulaRequestParameter
	if dataRequestParam, ok := data.Get("request_parameter").(*schema.Set); ok {
		for _, requestParam := range dataRequestParam.List() {
			requestParamMap := requestParam.(map[string]interface{})
			requestParam := *basistheory.NewReactorFormulaRequestParameter(requestParamMap["name"].(string), requestParamMap["type"].(string))
			requestParam.SetDescription(requestParamMap["description"].(string))
			requestParam.SetOptional(requestParamMap["optional"].(bool))
			requestParams = append(requestParams, requestParam)
		}
	}

	reactorFormula.SetRequestParameters(requestParams)

	return reactorFormula
}

func flattenReactorFormulaConfigurationData(configurationOptions []basistheory.ReactorFormulaConfiguration) []interface{} {
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

func flattenReactorFormulaRequestParameterData(requestParameters []basistheory.ReactorFormulaRequestParameter) []interface{} {
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
