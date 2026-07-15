import Sortable from 'sortablejs'

export function createTodoSortable(element, options) {
  return Sortable.create(element, options)
}
