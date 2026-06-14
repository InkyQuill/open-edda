import { AlertCircle, CheckCircle2, Loader2, Server, Settings2 } from "lucide-react";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";

import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../../shared/ui/tabs";
import { loadModelVariants, loadProviderConfigs } from "./modelSettingsThunks";
import { modelSettingsActions } from "./modelSettingsSlice";

export function ModelSettingsPanel() {
  const dispatch = useDispatch<AppDispatch>();
  const {
    activeModelVariantId,
    error,
    modelVariantsByProviderId,
    modelsStatus,
    providers,
    providersStatus,
    selectedProviderId,
    loadingModelProviderId,
  } = useSelector((state: RootState) => state.modelSettings);
  const selectedProvider = providers.find((provider) => provider.id === selectedProviderId) ?? null;
  const selectedProviderModels = selectedProviderId ? (modelVariantsByProviderId[selectedProviderId] ?? []) : [];
  const selectedModel = selectedProviderModels.find((model) => model.id === activeModelVariantId) ?? null;

  useEffect(() => {
    if (providersStatus === "idle") {
      void dispatch(loadProviderConfigs());
    }
  }, [dispatch, providersStatus]);

  useEffect(() => {
    if (!selectedProviderId) return;
    if (modelVariantsByProviderId[selectedProviderId] || loadingModelProviderId === selectedProviderId) return;
    void dispatch(loadModelVariants({ providerId: selectedProviderId }));
  }, [dispatch, loadingModelProviderId, modelVariantsByProviderId, selectedProviderId]);

  function handleProviderSelect(providerId: string): void {
    dispatch(modelSettingsActions.setSelectedProviderId(providerId));
    if (!modelVariantsByProviderId[providerId] && loadingModelProviderId !== providerId) {
      void dispatch(loadModelVariants({ providerId }));
    }
  }

  function handleRefresh(): void {
    void dispatch(loadProviderConfigs());
    if (selectedProviderId) {
      void dispatch(loadModelVariants({ providerId: selectedProviderId }));
    }
  }

  return (
    <section className="flex h-full min-h-0 flex-col gap-4" aria-labelledby="model-settings-title">
      <header className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 id="model-settings-title" className="text-base font-semibold text-foreground">
            Model settings
          </h2>
          <p className="text-sm text-muted-foreground">
            {selectedProvider ? selectedProvider.name : "No provider selected"}
          </p>
        </div>
        <Button type="button" variant="outline" size="icon-sm" aria-label="Refresh model settings" onClick={handleRefresh}>
          <Settings2 />
        </Button>
      </header>

      <div className="grid gap-2 rounded-md border border-border bg-muted/30 p-3 text-sm">
        <div className="flex items-center justify-between gap-3">
          <span className="text-muted-foreground">Provider</span>
          <span className="truncate font-medium text-foreground">{selectedProvider?.name ?? "No provider selected"}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <span className="text-muted-foreground">Model</span>
          <span className="truncate font-medium text-foreground">
            {selectedModel ? `${selectedModel.name} (${selectedModel.model})` : "No model selected"}
          </span>
        </div>
      </div>

      {error ? (
        <p role="alert" className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          <AlertCircle className="mt-0.5 size-4" aria-hidden="true" />
          {error}
        </p>
      ) : null}

      <Tabs defaultValue="providers" className="min-h-0">
        <TabsList>
          <TabsTrigger value="providers">Providers</TabsTrigger>
          <TabsTrigger value="models">Models</TabsTrigger>
        </TabsList>

        <TabsContent value="providers" className="min-h-0">
          <div className="flex max-h-80 flex-col gap-2 overflow-auto pr-1">
            {providersStatus === "pending" ? (
              <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                Loading providers...
              </p>
            ) : null}

            {providersStatus !== "pending" && providers.length === 0 ? (
              <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                No providers configured.
              </p>
            ) : null}

            {providers.map((provider) => (
              <Button
                key={provider.id}
                type="button"
                variant={provider.id === selectedProviderId ? "secondary" : "outline"}
                className="h-auto justify-start px-3 py-2 text-left"
                aria-pressed={provider.id === selectedProviderId}
                onClick={() => handleProviderSelect(provider.id)}
              >
                <Server data-icon="inline-start" aria-hidden="true" />
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium">{provider.name}</span>
                  <span className="block truncate text-xs text-muted-foreground">{provider.baseUrl}</span>
                </span>
              </Button>
            ))}
          </div>
        </TabsContent>

        <TabsContent value="models" className="min-h-0">
          <div className="flex max-h-80 flex-col gap-2 overflow-auto pr-1">
            {!selectedProviderId ? (
              <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                Select a provider before choosing a model.
              </p>
            ) : null}

            {selectedProviderId && modelsStatus === "pending" ? (
              <p className="flex items-center gap-2 rounded-md border border-border p-3 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" aria-hidden="true" />
                Loading models...
              </p>
            ) : null}

            {selectedProviderId && modelsStatus !== "pending" && selectedProviderModels.length === 0 ? (
              <p className="rounded-md border border-dashed border-border p-3 text-sm text-muted-foreground">
                No models configured for this provider.
              </p>
            ) : null}

            {selectedProviderModels.map((model) => (
              <Button
                key={model.id}
                type="button"
                variant={model.id === activeModelVariantId ? "secondary" : "outline"}
                className="h-auto justify-start px-3 py-2 text-left"
                aria-pressed={model.id === activeModelVariantId}
                onClick={() => dispatch(modelSettingsActions.setActiveModelVariantId(model.id))}
              >
                {model.id === activeModelVariantId ? (
                  <CheckCircle2 data-icon="inline-start" aria-hidden="true" />
                ) : (
                  <span data-icon="inline-start" className="size-4" aria-hidden="true" />
                )}
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium">{model.name}</span>
                  <span className="block truncate text-xs text-muted-foreground">{model.model}</span>
                </span>
              </Button>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </section>
  );
}
