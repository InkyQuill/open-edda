import { AppearanceProvider } from "../../features/appearance/Appearance";
import type { ReactNode } from "react";
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import { store } from "../store/store";

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <Provider store={store}>
      <AppearanceProvider><BrowserRouter>{children}</BrowserRouter></AppearanceProvider>
    </Provider>
  );
}
