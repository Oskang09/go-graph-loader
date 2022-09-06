export default {
    introspection: {
        type: 'file',
        location: 'schema/schema.json',
    },
    website: {
        template: 'carbon-multi-page',
        options: {
            queryGenerationFactories: {
                'GoMap': '{}',
                'GoArray': '[]',
                'RawString': '',
                'GoStringer': '',
            }
        }
    },
}