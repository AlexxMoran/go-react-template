import type { TApiResponseWrapper } from "@shared/services/api/base-api-service/types";
import type { IEntityIdField } from "@shared/types/commonEntity.types";
import type { IPaginationMeta, IPaginationParams } from "@shared/types/pagination.types";

export interface IPaginationServiceParams<TEntity extends IEntityIdField, TParams extends IPaginationParams> {
  loadFn: (params: TParams) => Promise<{ data: TApiResponseWrapper<TEntity[]> & IPaginationMeta }>;
  initImmediately?: boolean;
}
