import { describe, expect, it } from 'vitest'
import { pipelineIdentity, swapInSequence } from './pipeline-order'

describe('pipelineIdentity', () => {
  it('is the file name without directory or extension', () => {
    expect(pipelineIdentity('custom/edr-1234.yaml')).toBe('edr-1234')
    expect(pipelineIdentity('aws.yml')).toBe('aws')
  })

  it('ignores the suffix a deactivated file carries', () => {
    expect(pipelineIdentity('linux.yaml.disabled')).toBe('linux')
  })
})

describe('swapInSequence', () => {
  const all = ['aws', 'azure', 'custom', 'linux']

  it('exchanges the two and leaves the rest in place', () => {
    expect(swapInSequence(all, 'azure', 'linux')).toEqual(['aws', 'linux', 'custom', 'azure'])
  })

  it('keeps the whole sequence when the view showed only some of it', () => {
    const swapped = swapInSequence(all, 'custom', 'aws')
    expect(swapped).toHaveLength(all.length)
    expect(swapped).toEqual(['custom', 'azure', 'aws', 'linux'])
  })

  it('does not touch the input', () => {
    swapInSequence(all, 'aws', 'azure')
    expect(all).toEqual(['aws', 'azure', 'custom', 'linux'])
  })

  it('gives up when one of them is gone', () => {
    expect(swapInSequence(all, 'aws', 'deleted')).toBeNull()
  })
})
