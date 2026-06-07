import type { RootService } from "@shared/services/root-service";
import { createContext, useContext } from "react";

export const RootServiceContext = createContext({} as RootService);
export const useRootService = () => useContext(RootServiceContext);
