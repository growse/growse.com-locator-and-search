module.exports = function (grunt) {
    grunt.initConfig({
        jshint: {
            files: ['Gruntfile.js', 'js/scripts.js'],
            options: {
                globals: {
                    jQuery: true
                }
            }
        },
        watch: {
            scripts: {
                files: ['<%= jshint.files %>'],
                tasks: ['jshint', 'uglify', 'clean:js', 'rev']
            },
            stylesheets: {
                files: ['css/style.scss', 'css/solarizeddark.scss', 'css/nanoscroller.css'],
                tasks: ['sass', 'cssmin', 'clean:css', 'rev']
            }
        },
        uglify: {
            target: {
                files: {
                    'js/www.js': ['js/jquery-2.1.3.min.js', 'js/jquery.timeago.js', 'js/jquery.nanoscroller.min.js', 'js/highlight.pack.js', 'js/scripts.js'],
                    'js/where.js': ['js/d3.js', 'js/topojson.v1.min.js'],
                    'js/admin.js': ['js/showdown.js']
                }
            },
            options: {
                mangle: false,
                compress: false,
                beautify: false,
                sourceMap: true
            }
        },
        rev: {
            options: {
                algorithm: 'md5',
                length: 16
            },
            assets: {
                files: [{
                    src: ['js/www.js', 'js/where.js', 'js/admin.js', 'css/www.css']
                }]
            }
        },
        sass: {
            target: {
                files: {
                    'css/style.css': 'css/style.scss'
                }
            }
        },
        cssmin: {
            combine: {
                files: {
                    'css/www.css': ['css/style.css', 'css/solarizeddark.css', 'css/nanoscroller.css']
                }
            }
        },
        clean: {
            css: [
                'css/*.www.css'
            ],
            js: [
                'js/*.www.js',
                'js/*.admin.js',
                'js/*.where.js'
            ]
        }
    });

    grunt.loadNpmTasks('grunt-contrib-jshint');
    grunt.loadNpmTasks('grunt-contrib-watch');
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-rev');
    grunt.loadNpmTasks('grunt-sass');
    grunt.loadNpmTasks('grunt-contrib-cssmin');
    grunt.loadNpmTasks('grunt-contrib-clean');

    grunt.registerTask('default', ['jshint', 'uglify', 'sass', 'cssmin', 'clean', 'rev']);

};
