import { DEFAULT_PAGINATION_LIMIT } from "@shared/services/pagination-service/constants";
import type { IPaginationServiceParams } from "@shared/services/pagination-service/types";
import type { IEntityIdField, TEntityId } from "@shared/types/commonEntity.types";
import type { TMaybe } from "@shared/types/main.types";
import type { IPaginationParams } from "@shared/types/pagination.types";
import { makeAutoObservable, runInAction } from "mobx";

export class PaginationService<TEntity extends IEntityIdField, TParams extends IPaginationParams> {
  private listMap = new Map<TEntityId, TEntity>();

  limit = DEFAULT_PAGINATION_LIMIT;
  filteredCount: TMaybe<number> = null;
  totalCount: TMaybe<number> = null;
  isPaginating = false;
  isInitialLoading = false;
  hasMore = false;

  constructor(private params: IPaginationServiceParams<TEntity, TParams>) {
    makeAutoObservable(this);

    if (params.initImmediately) {
      void this.init();
    }
  }

  private get paginationParams() {
    return { skip: this.list.length, limit: this.limit };
  }

  get list() {
    return Array.from(this.listMap.values());
  }

  private setList = (list: TEntity[]) => {
    list.forEach((entity) => this.setItem(entity));
  };

  setItem = (entity: TEntity) => {
    this.listMap.set(entity.id, entity);
  };

  paginate = async (params?: Partial<TParams>) => {
    if (this.isPaginating || this.isInitialLoading || !this.hasMore) {
      return;
    }

    try {
      this.isPaginating = true;
      const { data } = await this.params.loadFn({
        ...(this.paginationParams as TParams),
        ...params
      });

      runInAction(() => {
        this.setList(data.data);
        this.totalCount = data.total_count;
        this.filteredCount = data.filtered_count;
        this.hasMore = this.list.length < data.filtered_count;
      });
    } finally {
      runInAction(() => {
        this.isPaginating = false;
      });
    }
  };

  init = async (params?: Partial<TParams>) => {
    this.listMap.clear();
    this.isPaginating = false;
    this.isInitialLoading = true;
    this.filteredCount = null;
    this.totalCount = null;

    try {
      const { data } = await this.params.loadFn({
        ...(this.paginationParams as TParams),
        ...params
      });

      runInAction(() => {
        this.setList(data.data);
        this.totalCount = data.total_count;
        this.filteredCount = data.filtered_count;
        this.hasMore = this.list.length < data.filtered_count;
      });
    } finally {
      runInAction(() => {
        this.isInitialLoading = false;
      });
    }
  };
}
