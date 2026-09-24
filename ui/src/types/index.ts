export interface ImageReference {
  original: string;
  registry?: string;
  repository?: string;
  tag?: string;
  digest?: string;
  architecture: string;
  os: string;
  source: string;
}

export interface LayerSummary {
  index: number;
  digest: string;
  command: string;
  size: number;
  wastedBytes: number;
  fileCount: number;
}

export interface ImageData {
  reference: ImageReference;
  totalSizeBytes: number;
  totalFilesCount: number;
  efficiencyScore: number;
  grade: string;
  layers: LayerSummary[];
  analyzedAt: string;
}

export type ChangeType = 'added' | 'modified' | 'deleted' | 'unchanged';

export interface VFSNode {
  path: string;
  name: string;
  size: number;
  isDir: boolean;
  changeType: ChangeType;
  isWasted: boolean;
  wastedBytes: number;
  wasteReason?: string;
  children?: string[];
}

export interface LayerTreeResponse {
  layerIndex: number;
  command: string;
  totalNodes: number;
  nodes: Record<string, VFSNode>;
}

export interface SecretFinding {
  id: string;
  ruleName: string;
  severity: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW';
  filePath: string;
  layerIndex: number;
  command: string;
  lineNumber: number;
  matchMasked: string;
  description: string;
  isDeletedInFinal: boolean;
}

export interface SecurityReport {
  totalFindings: number;
  criticalCount: number;
  highCount: number;
  mediumCount: number;
  lowCount: number;
  intermediateLeaks: number;
  runsAsRoot: boolean;
  suidBinariesCount: number;
  findings: SecretFinding[];
}

export interface Vulnerability {
  id: string;
  summary: string;
  details?: string;
  severity: string;
  score?: number;
  fixedIn?: string[];
  aliases?: string[];
}

export interface SBOMPackage {
  name: string;
  version: string;
  type: string;
  license?: string;
  description?: string;
  purl?: string;
  size?: number;
  vulnerabilities?: Vulnerability[];
}

export interface SBOMReport {
  imageName: string;
  totalPackages: number;
  osPackages: number;
  appPackages: number;
  totalVulnerabilities?: number;
  packages: SBOMPackage[];
}

export interface Recommendation {
  id: string;
  category: string;
  title: string;
  severity: 'high' | 'medium' | 'low' | 'info';
  estimatedSavingsBytes: number;
  explanation: string;
  remediationSnippet?: string;
  targetLayerIndex?: number;
}

export interface AdvisorReport {
  efficiencyScore: number;
  grade: string;
  totalImageBytes: number;
  totalWastedBytes: number;
  potentialSavingsBytes: number;
  recommendations: Recommendation[];
}

export interface FileDiff {
  path: string;
  status: 'added' | 'removed' | 'modified' | 'same';
  sizeA: number;
  sizeB: number;
  sizeDelta: number;
  isDir: boolean;
}

export interface DiffReport {
  imageAName: string;
  imageBName: string;
  totalSizeA: number;
  totalSizeB: number;
  totalSizeDelta: number;
  addedCount: number;
  removedCount: number;
  modifiedCount: number;
  sameCount: number;
  files: FileDiff[];
}

export interface FilePreview {
  path: string;
  name: string;
  size: number;
  mode: string;
  modTime: string;
  isDir: boolean;
  isSymlink: boolean;
  linkTarget?: string;
  digest?: string;
  layerIndex: number;
  changeType: ChangeType;
  isWasted: boolean;
  wastedBytes: number;
  wasteReason?: string;
  isText: boolean;
  content: string;
  lineCount: number;
  truncated: boolean;
  mimeType: string;
}

export interface DockerfileOptimization {
  originalCommands: string[];
  optimizedDockerfile: string;
  generatedDockerignore: string;
  improvements: string[];
  estimatedSavingsMB: number;
}

export interface ImagePreset {
  id: string;
  name: string;
  target: string;
  description: string;
  isDemo: string;
}
