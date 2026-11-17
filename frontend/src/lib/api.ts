import { ConnectionService, MetadataService, QueryService } from "../../bindings/github.com/collins/GoDBClient/internal/bridge";
import type * as models from "../../bindings/github.com/collins/GoDBClient/internal/models/models";
import { Duration } from "../../bindings/time/models";

export async function getConnections(): Promise<models.Profile[]> {
  const profiles = await ConnectionService.ListProfiles();
  return (profiles ?? []).filter((p): p is models.Profile => p !== null);
}
export async function saveConnection(profile: models.Profile, password: string): Promise<void> {
  return await ConnectionService.SaveProfile(profile, password);
}
export async function testConnection(profile: models.Profile, password = ""): Promise<void> {
  return await ConnectionService.TestConnection(profile, password);
}
export async function connect(id: string): Promise<void> { return await ConnectionService.Connect(id); }
export async function disconnect(id: string): Promise<void> { return await ConnectionService.Disconnect(id); }
export async function listActive(): Promise<string[]> { return (await ConnectionService.ListActive()) ?? []; }

export async function getDatabases(connID: string): Promise<string[]> { return (await MetadataService.Databases(connID)) ?? []; }
export async function getTables(connID: string, schema: string): Promise<models.TableInfo[]> { return (await MetadataService.Tables(connID, schema)) ?? []; }
export async function getColumns(connID: string, schema: string, table: string): Promise<models.ColumnInfo[]> { return (await MetadataService.Columns(connID, schema, table)) ?? []; }
export async function getForeignKeys(connID: string, schema: string, table: string): Promise<models.ForeignKeyInfo[]> { return (await MetadataService.ForeignKeys(connID, schema, table)) ?? []; }
export async function getIndexes(connID: string, schema: string, table: string): Promise<string[]> { return (await MetadataService.Indexes(connID, schema, table)) ?? []; }

export async function executeQuery(req: models.QueryRequest): Promise<models.Result | null> {
  return await QueryService.Execute(req);
}
export async function executeSql(connID: string, sql: string, page = 1, pageSize = 100, confirm = false): Promise<models.Result | null> {
  return await executeQuery({
    ConnectionID: connID,
    SQL: sql,
    Page: page,
    PageSize: pageSize,
    Timeout: Duration.Second * 30,
    Confirm: confirm,
  } as models.QueryRequest);
}
export async function inspectSql(sql: string): Promise<models.GuardDecision | null> {
  return await QueryService.Inspect(sql);
}

// legacy aliases
export const listProfiles = getConnections;
export const listDatabases = getDatabases;
export const listTables = getTables;
export const listColumns = getColumns;
export const saveProfile = saveConnection;
export const deleteProfile = (id: string) => ConnectionService.DeleteProfile(id);
export const cancelQuery = (connID: string) => QueryService.Cancel(connID);
