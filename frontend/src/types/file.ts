export interface MetaObject {
  id: number
  parent_id: number | null
  owner_id: number
  object_key: string | null
  file_type: string
  is_file: boolean | null
  name: string
  version: number
  created_at: string
  deleted_at: string | null
}
