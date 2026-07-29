export interface AddressRequest {
  address: string
  source_code?: string
}

export type QualityStatus =
  | "VALID"
  | "INCOMPLETE"
  | "AMBIGUOUS"
  | "CONFLICT"
  | "UNKNOWN"

export interface Location {
  province: string
  city: string
  district: string
  sub_district: string
  postal_code: string
}

export interface Conflict {
  type: string
  message: string
}

export interface Assessment {
  matched: string[]
  missing: string[]
  conflicts: Conflict[]
  ambiguous: string[]
}

export interface ResolutionCandidate {
  uuid: string
  score: number
  location: Location
  reasons: string[]
}

export interface Resolution {
  strategy: string[]
  candidate_count: number
  candidates: ResolutionCandidate[]
}

export interface Metadata {
  location_source: string
  location_version: string
}

export interface ResponseData {
  address_id: string
  status: QualityStatus
  confidence: number
  raw_input: string
  normalized_input: string
  formatted_address: string
  location: Location
  assessment: Assessment
  resolution: Resolution
  metadata: Metadata
}

export interface AddressResponse {
  timestamp: string
  request_id: string
  data: ResponseData
}

export interface ErrorResponse {
  timestamp: string
  request_id: string
  error: string
}
