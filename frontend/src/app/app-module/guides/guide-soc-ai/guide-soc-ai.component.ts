import {ChangeDetectorRef, Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';
import {UtmModuleGroupService} from '../../shared/services/utm-module-group.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ModuleChangeStatusBehavior} from '../../shared/behavior/module-change-status.behavior';

interface ProviderConfig {
  id: string;
  name: string;
  logo: string;
  description: string;
  fields: ProviderField[];
}

interface ProviderField {
  key: string;
  label: string;
  type: 'text' | 'password' | 'select' | 'number' | 'toggle' | 'headers';
  required?: boolean;
  placeholder?: string;
  options?: {value: string, label: string}[];
  tooltip?: string;
}

@Component({
  selector: 'app-guide-soc-ai',
  templateUrl: './guide-soc-ai.component.html',
  styleUrls: ['./guide-soc-ai.component.css']
})
export class GuideSocAiComponent implements OnInit {
  @Input() integrationId: number;
  @Input() serverId: number;
  module = UtmModulesEnum;

  activeProvider = 'openai';
  saving = false;
  loading = true;

  // Form values - what the user sees/edits
  formValues: {[key: string]: string} = {};
  customModelValue = '';

  // Editable headers table for custom provider
  headerRows: {key: string, value: string}[] = [];

  // Cache form values per provider so switching tabs preserves edits
  private providerFormCache: {[providerId: string]: {values: {[key: string]: string}, customModel: string, headers: {key: string, value: string}[]}} = {};

  // Raw configs from DB (to get IDs for saving)
  private rawConfigs: UtmModuleGroupConfType[] = [];
  private groupId: number;

  providers: ProviderConfig[] = [
    {
      id: 'openai',
      name: 'OpenAI',
      logo: 'assets/img/guides/soc-ai/providers/openai.svg',
      description: 'Obtain an API key from your OpenAI account. If you need help, see <a href="https://help.openai.com/en/articles/4936850-where-do-i-find-my-openai-api-key" target="_blank">Where do I find my API key?</a>',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true, placeholder: 'sk-...'},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'gpt-5.4', label: 'gpt-5.4'},
          {value: 'gpt-5.4-mini', label: 'gpt-5.4-mini'},
          {value: 'gpt-5.4-nano', label: 'gpt-5.4-nano'},
          {value: 'gpt-4.1', label: 'gpt-4.1'},
          {value: 'gpt-4.1-mini', label: 'gpt-4.1-mini'},
          {value: 'gpt-4.1-nano', label: 'gpt-4.1-nano'},
          {value: 'gpt-4o', label: 'gpt-4o'},
          {value: 'gpt-4o-mini', label: 'gpt-4o-mini'},
          {value: 'o3-mini', label: 'o3-mini'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle', tooltip: 'Automatically send new alerts for AI analysis'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle', tooltip: 'Automatically create incidents based on AI analysis'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle', tooltip: 'Allow SOC AI to update alert status after analysis'},
      ]
    },
    {
      id: 'anthropic',
      name: 'Anthropic',
      logo: 'assets/img/guides/soc-ai/providers/anthropic.svg',
      description: 'Obtain an API key from <a href="https://console.anthropic.com/settings/keys" target="_blank">Anthropic Console</a>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true, placeholder: 'sk-ant-...'},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'claude-sonnet-4-20250514', label: 'claude-sonnet-4-20250514'},
          {value: 'claude-haiku-4-20250414', label: 'claude-haiku-4-20250414'},
          {value: 'claude-opus-4-20250514', label: 'claude-opus-4-20250514'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'maxTokens', label: 'Max Tokens', type: 'number', required: true, placeholder: '4096', tooltip: 'Required for Anthropic'},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'azure',
      name: 'Azure OpenAI',
      logo: 'assets/img/guides/soc-ai/providers/azure.svg',
      description: 'Create an Azure OpenAI resource in the <a href="https://portal.azure.com/" target="_blank">Azure Portal</a>, deploy a model, and copy your API key from <strong>Keys and Endpoint</strong>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true},
        {key: 'url', label: 'Endpoint URL', type: 'text', required: true, placeholder: 'https://YOUR-RESOURCE.openai.azure.com/openai/deployments/YOUR-DEPLOYMENT/chat/completions?api-version=2024-02-01'},
        {key: 'model', label: 'Model', type: 'text', required: true, placeholder: 'gpt-4o'},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'gemini',
      name: 'Google Gemini',
      logo: 'assets/img/guides/soc-ai/providers/gemini.svg',
      description: 'Obtain an API key from <a href="https://aistudio.google.com/apikey" target="_blank">Google AI Studio</a>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'gemini-2.5-flash', label: 'gemini-2.5-flash'},
          {value: 'gemini-2.5-pro', label: 'gemini-2.5-pro'},
          {value: 'gemini-2.0-flash', label: 'gemini-2.0-flash'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'ollama',
      name: 'Ollama',
      logo: 'assets/img/guides/soc-ai/providers/ollama.svg',
      description: 'Install Ollama from <a href="https://ollama.ai" target="_blank">ollama.ai</a> and pull a model (<code>ollama pull llama3</code>). No API key required.',
      fields: [
        {key: 'url', label: 'Ollama Server URL', type: 'text', required: true, placeholder: 'http://YOUR-SERVER:11434/v1/chat/completions'},
        {key: 'model', label: 'Model', type: 'text', required: true, placeholder: 'llama3'},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'mistral',
      name: 'Mistral AI',
      logo: 'assets/img/guides/soc-ai/providers/mistral.svg',
      description: 'Obtain an API key from <a href="https://console.mistral.ai/api-keys/" target="_blank">Mistral Console</a>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'mistral-large-latest', label: 'mistral-large-latest'},
          {value: 'mistral-small-latest', label: 'mistral-small-latest'},
          {value: 'mistral-medium-latest', label: 'mistral-medium-latest'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'deepseek',
      name: 'DeepSeek',
      logo: 'assets/img/guides/soc-ai/providers/deepseek.svg',
      description: 'Obtain an API key from <a href="https://platform.deepseek.com/api_keys" target="_blank">DeepSeek Platform</a>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'deepseek-chat', label: 'deepseek-chat'},
          {value: 'deepseek-reasoner', label: 'deepseek-reasoner'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'groq',
      name: 'Groq',
      logo: 'assets/img/guides/soc-ai/providers/groq.svg',
      description: 'Obtain an API key from <a href="https://console.groq.com/keys" target="_blank">Groq Console</a>.',
      fields: [
        {key: 'apiKey', label: 'API Key', type: 'password', required: true},
        {key: 'model', label: 'Model', type: 'select', required: true, options: [
          {value: 'llama-3.3-70b-versatile', label: 'llama-3.3-70b-versatile'},
          {value: 'llama-3.1-8b-instant', label: 'llama-3.1-8b-instant'},
          {value: 'mixtral-8x7b-32768', label: 'mixtral-8x7b-32768'},
          {value: '__custom__', label: 'Custom model...'},
        ]},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
    {
      id: 'custom',
      name: 'Custom',
      logo: 'assets/img/guides/soc-ai/providers/custom.svg',
      description: 'Configure any provider that supports the OpenAI-compatible <code>/chat/completions</code> format.',
      fields: [
        {key: 'url', label: 'API URL', type: 'text', required: true, placeholder: 'https://your-provider.com/v1/chat/completions'},
        {key: 'model', label: 'Model', type: 'text', required: true, placeholder: 'model-name'},
        {key: 'authType', label: 'Auth Type', type: 'select', options: [
          {value: 'custom-headers', label: 'Custom Headers'},
          {value: 'none', label: 'None'},
        ]},
        {key: 'customHeaders', label: 'Custom Headers', type: 'headers'},
        {key: 'maxTokens', label: 'Max Tokens', type: 'number', placeholder: '4096'},
        {key: 'autoAnalyze', label: 'Auto-analyze alerts', type: 'toggle'},
        {key: 'incidentCreation', label: 'Auto-create incidents', type: 'toggle'},
        {key: 'changeAlertStatus', label: 'Change alert status after analysis', type: 'toggle'},
      ]
    },
  ];

  // Map provider to their auth header format
  private providerAuthHeaders: {[providerId: string]: {headerName: string, headerValuePrefix: string}} = {
    'openai': {headerName: 'Authorization', headerValuePrefix: 'Bearer '},
    'anthropic': {headerName: 'x-api-key', headerValuePrefix: ''},
    'azure': {headerName: 'api-key', headerValuePrefix: ''},
    'gemini': {headerName: 'Authorization', headerValuePrefix: 'Bearer '},
    'mistral': {headerName: 'Authorization', headerValuePrefix: 'Bearer '},
    'deepseek': {headerName: 'Authorization', headerValuePrefix: 'Bearer '},
    'groq': {headerName: 'Authorization', headerValuePrefix: 'Bearer '},
  };

  // Map from our form keys to backend confKeys
  private keyMap = {
    'model': 'utmstack.socai.model',
    'url': 'utmstack.socai.url',
    'maxTokens': 'utmstack.socai.maxTokens',
    'authType': 'utmstack.socai.authType',
    'customHeaders': 'utmstack.socai.customHeaders',
    'autoAnalyze': 'utmstack.socai.autoAnalyze',
    'incidentCreation': 'utmstack.socai.incidentCreation',
    'changeAlertStatus': 'utmstack.socai.changeAlertStatus',
  };

  constructor(
    private moduleGroupService: UtmModuleGroupService,
    private moduleGroupConfService: UtmModuleGroupConfService,
    private toast: UtmToastService,
    private moduleChangeStatusBehavior: ModuleChangeStatusBehavior,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    this.loadConfig();
  }

  get currentProvider(): ProviderConfig {
    return this.providers.find(p => p.id === this.activeProvider);
  }

  onTabChange(event: any) {
    // Save current form state
    this.providerFormCache[this.activeProvider] = {
      values: {...this.formValues},
      customModel: this.customModelValue,
      headers: [...this.headerRows.map(r => ({...r}))]
    };

    this.activeProvider = event.nextId;

    // Restore from cache or init fresh
    const cached = this.providerFormCache[this.activeProvider];
    if (cached) {
      this.formValues = {...cached.values};
      this.customModelValue = cached.customModel;
      this.headerRows = cached.headers.map(r => ({...r}));
    } else {
      this.initFormValues();
    }
  }

  getModelValue(): string {
    if (this.formValues['model'] === '__custom__') {
      return this.customModelValue;
    }
    return this.formValues['model'] || '';
  }

  isFormValid(): boolean {
    const provider = this.currentProvider;
    return provider.fields
      .filter(f => f.required)
      .every(f => {
        if (f.key === 'model') {
          return !!this.getModelValue();
        }
        const val = this.formValues[f.key];
        return val !== undefined && val !== null && val !== '';
      });
  }

  private loadConfig() {
    this.loading = true;
    this.moduleGroupService.query({moduleId: this.integrationId}).subscribe(response => {
      const groups = response.body || [];
      if (groups.length > 0) {
        this.groupId = groups[0].id;
        this.rawConfigs = groups[0].moduleGroupConfigurations || [];
        this.loadCurrentProvider();
        this.initFormValues();
        // Cache the initial loaded values for the saved provider
        this.providerFormCache[this.activeProvider] = {
          values: {...this.formValues},
          customModel: this.customModelValue,
          headers: [...this.headerRows.map(r => ({...r}))]
        };
      }
      this.loading = false;
      this.cdr.detectChanges();
    }, () => {
      this.loading = false;
      this.cdr.detectChanges();
    });
  }

  private loadCurrentProvider() {
    const providerConf = this.rawConfigs.find(c => c.confKey === 'utmstack.socai.provider');
    if (providerConf && providerConf.confValue) {
      this.activeProvider = providerConf.confValue;
    }
    this.savedProvider = this.activeProvider;
  }

  private savedProvider: string;

  private initFormValues() {
    const isCurrentSavedProvider = this.activeProvider === this.savedProvider;

    // Always load behavior toggles from DB
    const autoAnalyze = this.getConf('utmstack.socai.autoAnalyze');
    const incidentCreation = this.getConf('utmstack.socai.incidentCreation');
    const changeStatus = this.getConf('utmstack.socai.changeAlertStatus');

    // Start clean
    this.formValues = {
      'autoAnalyze': autoAnalyze ? autoAnalyze.confValue : 'false',
      'incidentCreation': incidentCreation ? incidentCreation.confValue : 'false',
      'changeAlertStatus': changeStatus ? changeStatus.confValue : 'false',
    };
    this.customModelValue = '';

    // Only load provider-specific values if viewing the saved provider
    if (isCurrentSavedProvider) {
      const modelConf = this.getConf('utmstack.socai.model');
      if (modelConf && modelConf.confValue) {
        // Check if the model value matches any option in the current provider
        const provider = this.currentProvider;
        const modelField = provider.fields.find(f => f.key === 'model');
        if (modelField && modelField.options) {
          const match = modelField.options.find(o => o.value === modelConf.confValue);
          if (match) {
            this.formValues['model'] = modelConf.confValue;
          } else {
            this.formValues['model'] = '__custom__';
            this.customModelValue = modelConf.confValue;
          }
        } else {
          this.formValues['model'] = modelConf.confValue;
        }
      }

      const urlConf = this.getConf('utmstack.socai.url');
      if (urlConf) {
        this.formValues['url'] = urlConf.confValue || '';
      }

      // Check if API key exists in custom headers — show masked if so
      const customHeaders = this.getConf('utmstack.socai.customHeaders');
      if (customHeaders && customHeaders.confValue && customHeaders.confValue !== '{}') {
        try {
          const headers = JSON.parse(customHeaders.confValue);
          const authConfig = this.providerAuthHeaders[this.activeProvider];
          if (authConfig && headers[authConfig.headerName]) {
            // API key exists — show masked, don't expose the real value
            this.formValues['apiKey'] = '*****';
          }
        } catch (e) {}
        this.formValues['customHeaders'] = customHeaders.confValue;
        this.parseHeadersFromJson(customHeaders.confValue);
      }

      const maxTokensConf = this.getConf('utmstack.socai.maxTokens');
      if (maxTokensConf) {
        this.formValues['maxTokens'] = maxTokensConf.confValue || '';
      }

      const authType = this.getConf('utmstack.socai.authType');
      if (authType) {
        this.formValues['authType'] = authType.confValue || 'custom-headers';
      }

      const customHeaders = this.getConf('utmstack.socai.customHeaders');
      if (customHeaders) {
        this.formValues['customHeaders'] = customHeaders.confValue || '{}';
        this.parseHeadersFromJson(this.formValues['customHeaders']);
      }
    } else {
      this.headerRows = [];
      // Set defaults for the new provider
      this.setFieldDefaults();
    }
  }

  private setFieldDefaults() {
    const provider = this.currentProvider;
    for (const field of provider.fields) {
      if (field.type === 'select' && field.options && field.options.length > 0 && !this.formValues[field.key]) {
        this.formValues[field.key] = field.options[0].value;
      }
      if (field.key === 'maxTokens' && !this.formValues['maxTokens']) {
        this.formValues['maxTokens'] = '4096';
      }
    }
  }

  private getConf(confKey: string): UtmModuleGroupConfType {
    return this.rawConfigs.find(c => c.confKey === confKey);
  }

  save() {
    this.saving = true;
    const changes: UtmModuleGroupConfType[] = [];

    // Set provider
    this.pushChange(changes, 'utmstack.socai.provider', this.activeProvider);

    // Set model
    this.pushChange(changes, 'utmstack.socai.model', this.getModelValue());

    // Set URL for providers that need it (azure, ollama, custom)
    if (this.formValues['url']) {
      this.pushChange(changes, 'utmstack.socai.url', this.formValues['url']);
    }

    // Set maxTokens
    if (this.formValues['maxTokens']) {
      this.pushChange(changes, 'utmstack.socai.maxTokens', this.formValues['maxTokens']);
    }

    // Set behavior toggles
    this.pushChange(changes, 'utmstack.socai.autoAnalyze', this.formValues['autoAnalyze'] || 'false');
    this.pushChange(changes, 'utmstack.socai.incidentCreation', this.formValues['incidentCreation'] || 'false');
    this.pushChange(changes, 'utmstack.socai.changeAlertStatus', this.formValues['changeAlertStatus'] || 'false');

    // Build auth headers
    if (this.activeProvider === 'custom') {
      // Custom provider: user manages auth type and headers directly
      this.pushChange(changes, 'utmstack.socai.authType', this.formValues['authType'] || 'custom-headers');
      this.pushChange(changes, 'utmstack.socai.customHeaders', this.formValues['customHeaders'] || '{}');
    } else if (this.activeProvider === 'ollama') {
      // Ollama: no auth needed
      this.pushChange(changes, 'utmstack.socai.authType', 'none');
      this.pushChange(changes, 'utmstack.socai.customHeaders', '{}');
    } else {
      // Known providers: build auth header from API key
      const authConfig = this.providerAuthHeaders[this.activeProvider];
      if (authConfig && this.formValues['apiKey'] && this.formValues['apiKey'] !== '*****') {
        // User entered a new API key — build auth headers
        const headers: {[k: string]: string} = {};
        headers[authConfig.headerName] = authConfig.headerValuePrefix + this.formValues['apiKey'];
        this.pushChange(changes, 'utmstack.socai.authType', 'custom-headers');
        this.pushChange(changes, 'utmstack.socai.customHeaders', JSON.stringify(headers));
      }
      // If apiKey is '*****', don't touch customHeaders — keep existing value in DB
    }

    this.moduleGroupConfService.update({
      keys: changes,
      moduleId: this.integrationId
    }).subscribe(
      () => {
        this.saving = false;
        this.toast.showSuccessBottom('SOC AI configuration saved successfully');
      },
      (err) => {
        this.saving = false;
        if (err.status === 400) {
          const message = this.extractValidationError(err);
          this.toast.showError('Invalid Configuration', message);
        } else {
          this.toast.showError('Error', 'Failed to save configuration. Please try again.');
        }
      }
    );
  }

  private pushChange(changes: UtmModuleGroupConfType[], confKey: string, value: string) {
    const existing = this.getConf(confKey);
    if (existing) {
      changes.push({
        ...existing,
        confValue: value,
        confOptions: existing.confOptions ? JSON.stringify(existing.confOptions) : existing.confOptions,
        confVisibility: existing.confVisibility ? JSON.stringify(existing.confVisibility) : existing.confVisibility,
      });
    }
  }

  private extractValidationError(err: any): string {
    const defaultMsg = 'The configuration data is invalid. Please check your inputs and try again.';
    try {
      const body = err.error;
      if (body && body.fieldErrors && body.fieldErrors.length > 0) {
        return body.fieldErrors.map((e: any) => e.message).join('. ');
      }
      if (body && body.message) {
        return body.message;
      }
      const headerError = err.headers ? err.headers.get('X-UtmStack-error') : null;
      if (headerError) {
        return headerError;
      }
      if (typeof body === 'string' && body.length > 0) {
        return body;
      }
    } catch (e) {}
    return defaultMsg;
  }

  onToggle(key: string, value: boolean) {
    this.formValues[key] = value.toString();
  }

  addHeaderRow() {
    this.headerRows.push({key: '', value: ''});
    this.syncHeadersToForm();
  }

  removeHeaderRow(index: number) {
    this.headerRows.splice(index, 1);
    this.syncHeadersToForm();
  }

  syncHeadersToForm() {
    const obj: {[k: string]: string} = {};
    for (const row of this.headerRows) {
      if (row.key.trim()) {
        obj[row.key.trim()] = row.value;
      }
    }
    this.formValues['customHeaders'] = JSON.stringify(obj);
  }

  private parseHeadersFromJson(json: string) {
    this.headerRows = [];
    try {
      const obj = JSON.parse(json || '{}');
      for (const key of Object.keys(obj)) {
        this.headerRows.push({key, value: obj[key]});
      }
    } catch (e) {
      // Invalid JSON, start empty
    }
  }
}
