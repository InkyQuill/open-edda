import { AlertCircle, Loader2, Settings2 } from "lucide-react";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";

import type { AppDispatch, RootState } from "../../app/store/store";
import { Button } from "../../shared/ui/button";
import { loadModelVariants, loadProviderConfigs } from "./modelSettingsThunks";

export function ModelStatus() {
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
  const isLoading = providersStatus === "pending" || modelsStatus === "pending";

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

  return (
    <section className="flex flex-col gap-3" aria-labelledby="model-status-title">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h3 id="model-status-title" className="text-sm font-medium text-foreground">
            Model
          </h3>
          <p className="text-xs text-muted-foreground">{selectedProvider?.name ?? "No provider selected"}</p>
        </div>
        <Button
          asChild
          type="button"
          variant="outline"
          size="icon-sm"
          aria-label="Open model settings"
        >
          <Link to="/settings">
            {isLoading ? <Loader2 className="animate-spin" /> : <Settings2 />}
          </Link>
        </Button>
      </div>

      <div className="flex flex-col gap-1 rounded-md border border-border bg-muted/40 p-3 text-sm">
        <p className="font-medium text-foreground">
          {selectedModel ? selectedModel.name : "No model selected"}
        </p>
        <p className="text-xs text-muted-foreground">
          {selectedModel ? selectedModel.model : "Assistant actions are disabled until a model is selected."}
        </p>
      </div>

      {!activeModelVariantId ? (
        <div className="flex items-start gap-2 rounded-md border border-border bg-background p-3 text-sm">
          <AlertCircle className="mt-0.5 size-4 text-muted-foreground" aria-hidden="true" />
          <div>
            <p className="font-medium text-foreground">AI actions disabled</p>
            <p className="text-xs text-muted-foreground">Select a provider and model before using assistant actions.</p>
          </div>
        </div>
      ) : null}

      {error ? (
        <div role="alert" className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
          <AlertCircle className="mt-0.5 size-4" aria-hidden="true" />
          <p>{error}</p>
        </div>
      ) : null}
    </section>
  );
}
