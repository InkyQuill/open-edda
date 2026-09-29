import type { ReactNode } from "react";
import { Provider } from "react-redux";
import { BrowserRouter } from "react-router-dom";
import { ThemeProvider } from "../../features/appearance/ThemeProvider";
import { store } from "../store/store";

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <Provider store={store}>
      <BrowserRouter><ThemeProvider>{children}</ThemeProvider></BrowserRouter>
    </Provider>
  );
}
