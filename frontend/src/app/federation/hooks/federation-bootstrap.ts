import {FederationModeService} from '../services/federation-mode.service';

export function initFederationMode(modeService: FederationModeService): () => Promise<void> {
  return () => new Promise<void>(resolve => {
    modeService.detect().subscribe({
      next: () => resolve(),
      error: () => resolve()
    });
  });
}
