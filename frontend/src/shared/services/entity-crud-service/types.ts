import type { TApiResponseWrapper } from "@shared/services/api/base-api-service/types";
import type { IEntityIdField, TEntityId } from "@shared/types/commonEntity.types";
import type { IPaginationMeta, IPaginationParams } from "@shared/types/pagination.types";

export interface IEntityCrudServiceParams<
  TEntity extends IEntityIdField,
  TGetListParams extends IPaginationParams,
  TCreateParams = never,
  TEditParams = never
> {
  getEntitiesFn: (
    params: TGetListParams
  ) => Promise<{ data: TApiResponseWrapper<TEntity[]> & IPaginationMeta }>;
  createEntityFn?: (params: TCreateParams) => Promise<unknown>;
  editEntityFn?: (
    entityId: TEntityId,
    params: TEditParams
  ) => Promise<{ data: TApiResponseWrapper<TEntity> }>;
  deleteEntityFn?: (entityId: TEntityId) => Promise<unknown>;
  hasFiltersReaction?: boolean;
}
