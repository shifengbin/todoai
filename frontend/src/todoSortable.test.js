import { describe, expect, it, vi } from 'vitest'
import { createTodoSortable } from './todoSortable'

const sortableCreateMock = vi.hoisted(() => vi.fn())

vi.mock('sortablejs', () => ({
  default: {
    create: sortableCreateMock
  }
}))

describe('createTodoSortable', () => {
  it('delegates sortable creation through a replaceable module boundary', () => {
    const element = document.createElement('div')
    const options = { handle: '.todo-drag-handle' }
    const instance = { destroy: vi.fn() }
    sortableCreateMock.mockReturnValue(instance)

    expect(createTodoSortable(element, options)).toBe(instance)
    expect(sortableCreateMock).toHaveBeenCalledWith(element, options)
  })
})
