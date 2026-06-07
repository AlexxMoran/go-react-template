import { makeAutoObservable } from "mobx";

export class FilterService<TFilters extends object> {
  private _filters: Partial<TFilters> = {};

  constructor() {
    makeAutoObservable(this);
  }

  get filters() {
    return { ...this._filters };
  }

  setFilters(newFilters: Partial<TFilters>) {
    this._filters = newFilters;
  }

  setFilter<K extends keyof TFilters>(key: K, value: TFilters[K]) {
    if (value) {
      this._filters[key] = value;
    } else {
      delete this._filters[key];
    }
  }
}
