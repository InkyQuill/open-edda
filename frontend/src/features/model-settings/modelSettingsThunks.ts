import { createAsyncThunk } from "@reduxjs/toolkit";

import { listModelVariants, listProviderConfigs } from "../../agentApi";

export const loadProviderConfigs = createAsyncThunk(
  "modelSettings/loadProviderConfigs",
  async () => listProviderConfigs(),
);

export const loadModelVariants = createAsyncThunk(
  "modelSettings/loadModelVariants",
  async ({ providerId }: { providerId: string }) => listModelVariants(providerId),
);
