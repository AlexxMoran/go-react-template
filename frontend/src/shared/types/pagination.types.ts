export interface IPaginationParams {
  skip: number;
  limit: number;
}

export interface IPaginationMeta {
  filtered_count: number;
  total_count: number;
}
