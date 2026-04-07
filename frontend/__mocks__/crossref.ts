export const mockOutgoingRef = {
  target_entry_id: 'ent_002',
  target_ref: 'E2',
  entry_code: 'E2',
  entry_title: 'Reactive State Management',
  research_code: 'R1',
  research_name: 'Vue Component Architecture',
  resolved: true,
}

export const mockOutgoingRefCrossResearch = {
  target_entry_id: 'ent_r2_001',
  target_ref: 'R2:E1',
  entry_code: 'E1',
  entry_title: 'Pinia Store Patterns',
  research_code: 'R2',
  research_name: 'State Management Patterns',
  resolved: true,
}

export const mockOutgoingRefUnresolved = {
  target_entry_id: null,
  target_ref: 'E99',
  entry_code: null,
  entry_title: null,
  research_code: null,
  research_name: null,
  resolved: false,
}

export const mockIncomingRef = {
  source_id: 'ent_004',
  source_type: 'entry',
  entry_code: 'E4',
  entry_title: 'Slots and Render Functions',
  research_code: 'R1',
  research_name: 'Vue Component Architecture',
}

export const mockOutgoingRefs = [mockOutgoingRef, mockOutgoingRefCrossResearch, mockOutgoingRefUnresolved]
export const mockIncomingRefs = [mockIncomingRef]
