import type { components } from "@shared/api/schema";

export interface IPaginationParams {
  skip: number;
  limit: number;
}

/** Paging metadata returned alongside a list (skip/limit/filtered_count/total_count). */
export type IPaginationMeta = components["schemas"]["PaginationMeta"];
