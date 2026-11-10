// Thin wrapper over the Wails-generated service bindings.
//
// The generated bindings return `| null` for slices; the wrapper normalises
// that away so the rest of the application can rely on plain arrays.
import * as svc from "../bindings/github.com/incurl/sumate/internal/service/service";
import type {
  Breadcrumb,
  Filter,
  MatchResult,
  Stats,
  TreeNode,
  Verse,
} from "./types";

const orEmpty = <T>(v: T[] | null | undefined): T[] => v ?? [];
const orZero = <T>(v: T | null | undefined): T => v as T;

export const api = {
  stats: async (): Promise<Stats> => orZero(await svc.Stats()),
  roots: async (): Promise<TreeNode[]> => orEmpty(await svc.Roots()),
  children: async (nodeID: string, path: Filter[]): Promise<TreeNode[]> =>
    orEmpty(await svc.Children(nodeID, path)),
  preview: async (path: Filter[], limit: number): Promise<Verse[]> =>
    orEmpty(await svc.Preview(path, limit)),
  breadcrumb: async (path: Filter[]): Promise<Breadcrumb[]> =>
    orEmpty(await svc.Breadcrumb(path)),
  match: async (path: Filter[], description: string): Promise<MatchResult> =>
    orZero(await svc.Match(path, description)),
  philosophicalReading: async (
    path: Filter[],
  ): Promise<import("./philosophy").Assessment[]> =>
    orEmpty(await svc.PhilosophicalReading(path)),
  productStance: async () => orZero(await svc.ProductStance()),
  verse: async (id: string): Promise<Verse> => orZero(await svc.Verse(id)),
  search: async (query: string, limit: number): Promise<Verse[]> =>
    orEmpty(await svc.Search(query, limit)),
  allVerses: async (): Promise<Verse[]> => orEmpty(await svc.AllVerses()),
};

export type Api = typeof api;
