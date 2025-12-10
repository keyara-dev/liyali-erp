'use server';

/**
 * GRN Server Actions
 * Handles saving quality issues to GRNs in localStorage
 * Maintains single source of truth pattern
 */

interface QualityIssue {
  id: string;
  itemId: string;
  description: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
}

interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  poNumber: string;
  status: 'DRAFT' | 'SUBMITTED' | 'CONFIRMED' | 'REJECTED';
  warehouseLocation: string;
  receivedDate: string;
  receivedBy: string;
  approvedBy?: string;
  items: Array<{
    id: string;
    itemNumber: number;
    description: string;
    poQuantity: number;
    receivedQuantity: number;
    unit: string;
    variance: number;
    damage: number;
    damageNotes?: string;
    condition: 'GOOD' | 'DAMAGED' | 'PARTIAL';
  }>;
  qualityIssues: QualityIssue[];
  notes?: string;
  currentStage: number;
  stageName: string;
  createdAt: string;
  updatedAt: string;
}

const STORAGE_KEY = 'app_grns';

/**
 * Get all GRNs from localStorage
 */
function getGRNs(): GoodsReceivedNote[] {
  try {
    if (typeof window === 'undefined') return [];
    const data = localStorage.getItem(STORAGE_KEY);
    return data ? JSON.parse(data) : [];
  } catch (error) {
    console.error('Error reading GRNs from storage:', error);
    return [];
  }
}

/**
 * Get a single GRN by ID
 */
function getGRNById(id: string): GoodsReceivedNote | null {
  const grns = getGRNs();
  return grns.find((grn) => grn.id === id) || null;
}

/**
 * Save a GRN to localStorage
 */
function saveGRN(grn: GoodsReceivedNote): GoodsReceivedNote {
  try {
    if (typeof window === 'undefined') return grn;

    const grns = getGRNs();
    const existingIndex = grns.findIndex((g) => g.id === grn.id);

    if (existingIndex >= 0) {
      grns[existingIndex] = {
        ...grn,
        updatedAt: new Date().toISOString(),
      };
    } else {
      grns.push({
        ...grn,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    }

    localStorage.setItem(STORAGE_KEY, JSON.stringify(grns));
    return grn;
  } catch (error) {
    console.error('Error saving GRN to storage:', error);
    throw new Error('Failed to save GRN');
  }
}

/**
 * Add a quality issue to a GRN
 * @param grnId - The GRN ID
 * @param issue - The quality issue to add (without id)
 * @returns The updated GRN
 */
export async function addQualityIssueToGRN(
  grnId: string,
  issue: Omit<QualityIssue, 'id'>
): Promise<GoodsReceivedNote> {
  try {
    // Get the GRN from storage
    const grn = getGRNById(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

    // Create new issue with unique ID
    const newIssue: QualityIssue = {
      id: `issue-${Date.now()}`,
      ...issue,
    };

    // Add issue to GRN
    const updatedGRN: GoodsReceivedNote = {
      ...grn,
      qualityIssues: [...grn.qualityIssues, newIssue],
      updatedAt: new Date().toISOString(),
    };

    // Save updated GRN to localStorage
    saveGRN(updatedGRN);

    return updatedGRN;
  } catch (error) {
    console.error('Error adding quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to add quality issue');
  }
}

/**
 * Get a GRN by ID
 * @param grnId - The GRN ID
 * @returns The GRN or null if not found
 */
export async function getGRNAction(grnId: string): Promise<GoodsReceivedNote | null> {
  try {
    return getGRNById(grnId);
  } catch (error) {
    console.error('Error fetching GRN:', error);
    throw error;
  }
}

/**
 * Remove a quality issue from a GRN
 * @param grnId - The GRN ID
 * @param issueId - The issue ID to remove
 * @returns The updated GRN
 */
export async function removeQualityIssueFromGRN(
  grnId: string,
  issueId: string
): Promise<GoodsReceivedNote> {
  try {
    const grn = getGRNById(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

    const updatedGRN: GoodsReceivedNote = {
      ...grn,
      qualityIssues: grn.qualityIssues.filter((issue) => issue.id !== issueId),
      updatedAt: new Date().toISOString(),
    };

    saveGRN(updatedGRN);

    return updatedGRN;
  } catch (error) {
    console.error('Error removing quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to remove quality issue');
  }
}

/**
 * Update a quality issue in a GRN
 * @param grnId - The GRN ID
 * @param issueId - The issue ID to update
 * @param updates - The fields to update
 * @returns The updated GRN
 */
export async function updateQualityIssueInGRN(
  grnId: string,
  issueId: string,
  updates: Partial<Omit<QualityIssue, 'id'>>
): Promise<GoodsReceivedNote> {
  try {
    const grn = getGRNById(grnId);

    if (!grn) {
      throw new Error(`GRN with ID ${grnId} not found`);
    }

    const updatedGRN: GoodsReceivedNote = {
      ...grn,
      qualityIssues: grn.qualityIssues.map((issue) =>
        issue.id === issueId
          ? {
              ...issue,
              ...updates,
            }
          : issue
      ),
      updatedAt: new Date().toISOString(),
    };

    saveGRN(updatedGRN);

    return updatedGRN;
  } catch (error) {
    console.error('Error updating quality issue:', error);
    throw error instanceof Error ? error : new Error('Failed to update quality issue');
  }
}
