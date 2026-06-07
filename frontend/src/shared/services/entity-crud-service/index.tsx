import type { IEntityCrudServiceParams } from "@shared/services/entity-crud-service/types";
import { FilterService } from "@shared/services/filter-service";
import { PaginationService } from "@shared/services/pagination-service";
import type { IEntityIdField, TEntityId } from "@shared/types/commonEntity.types";
import type { IPaginationParams } from "@shared/types/pagination.types";
import { reaction, type IReactionDisposer } from "mobx";

export class EntityCrudService<
  TEntity extends IEntityIdField,
  TGetListParams extends IPaginationParams,
  TCreateParams = never,
  TEditParams = never
> {
  private paginationService: PaginationService<TEntity, TGetListParams>;
  private filterService = new FilterService<TGetListParams>();
  reactionList: IReactionDisposer[] = [];

  constructor(private params: IEntityCrudServiceParams<TEntity, TGetListParams, TCreateParams, TEditParams>) {
    const paginationService = new PaginationService({
      loadFn: params.getEntitiesFn
    });

    this.paginationService = paginationService;

    if (params.hasFiltersReaction) {
      this.reactionList.push(
        reaction(
          () => this.filterService.filters,
          (filters) => {
            void paginationService.init(filters);
          }
        )
      );
    }

    void paginationService.init(this.filterService.filters);
  }

  get listData() {
    return {
      list: this.paginationService.list,
      isInitialLoading: this.paginationService.isInitialLoading,
      isPaginating: this.paginationService.isPaginating,
      totalCount: this.paginationService.totalCount,
      filteredCount: this.paginationService.filteredCount,
      hasMore: this.paginationService.hasMore
    };
  }

  get filters() {
    return this.filterService.filters;
  }

  editEntity = async (entityId: TEntityId, params: TEditParams) => {
    const result = await this.params.editEntityFn?.(entityId, params);
    if (result) {
      this.paginationService.setItem(result.data.data);
      return result.data;
    }
    return undefined;
  };

  paginate = async () => {
    await this.paginationService.paginate(this.filterService.filters);
  };

  setFilters = (filters: TGetListParams) => {
    this.filterService.setFilters(filters);
  };
}
