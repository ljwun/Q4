import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from 'prosemirror-state'

export interface FileHandlerOptions {
    onUpload: (file: File) => Promise<string | null>
}

export const FileHandler = Extension.create<FileHandlerOptions>({
    name: 'fileHandler',

    addProseMirrorPlugins() {
        return [
            new Plugin({
                key: new PluginKey('fileHandler'),
                props: {
                    handleDOMEvents: {
                        drop: (view, event) => {
                            const { state } = view
                            const { selection } = state
                            const { $anchor } = selection

                            const files = (event as DragEvent).dataTransfer?.files

                            if (files && files.length > 0) {
                                event.preventDefault()

                                const file = files[0]
                                const pos = view.posAtCoords({ left: event.clientX, top: event.clientY })?.pos || $anchor.pos

                                this.options.onUpload(file).then((url) => {
                                    if (url) {
                                        const node = view.state.schema.nodes.image.create({ src: url })
                                        const transaction = view.state.tr.insert(pos, node)
                                        view.dispatch(transaction)
                                    }
                                })

                                return true
                            }

                            return false
                        },
                        paste: (view, event) => {
                            const items = (event as ClipboardEvent).clipboardData?.items

                            if (items) {
                                for (const item of items) {
                                    if (item.type.indexOf('image') === 0) {
                                        event.preventDefault()

                                        const file = item.getAsFile()
                                        if (file) {
                                            this.options.onUpload(file).then((url) => {
                                                if (url) {
                                                    const node = view.state.schema.nodes.image.create({ src: url })
                                                    const transaction = view.state.tr.replaceSelectionWith(node)
                                                    view.dispatch(transaction)
                                                }
                                            })
                                        }

                                        return true
                                    }
                                }
                            }

                            return false
                        },
                    },
                },
            }),
        ]
    },
})

