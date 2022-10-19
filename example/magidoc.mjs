
		export default {
			introspection: {
				type: 'file',
				location: 'schema.json',
			},
			website: {
				template: 'carbon-multi-page',
				options: {
					queryGenerationFactories: {
						'RawString': '',
						'GoStringer': '',
						'gomap_string_string': '{}',
'gomap_string_interface': '{}',
'goslice_string': '[]',
'goarray_string': '[]'
					}
				}
			},
		}
	