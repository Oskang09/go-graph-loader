
		export default {
			introspection: {
				type: 'file',
				location: 'schema2/magidoc.mjs/schema.json',
			},
			website: {
				template: 'carbon-multi-page',
				options: {
					queryGenerationFactories: {
						'RawString': '',
						'GoStringer': '',
						'goarray_string': '[]',
'gomap_string_string': '{}',
'gomap_string_interface': '{}',
'goslice_string': '[]'
					}
				}
			},
		}
	