// These DTOs are aliases over the generated OpenAPI types (src/shared/api/schema).
// Regenerate with `npm run gen:api` after the backend contract changes — the TS
// compiler then flags every call site that no longer matches.
import type { components } from "@shared/api/schema";

export type IUserDto = components["schemas"]["User"];
export type ILoginDto = components["schemas"]["LoginRequest"];
export type IRegisterDto = components["schemas"]["RegisterRequest"];
export type IAccessToken = components["schemas"]["TokenResponse"];
export type IEditUserDto = components["schemas"]["UpdateProfileRequest"];
