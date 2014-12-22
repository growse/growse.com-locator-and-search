module.exports = function(grunt) {

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
            scripts:{
                files: ['<%= jshint.files %>'],
                tasks: ['jshint', 'uglify', 'rev']
            },
            stylesheets: {
                files: ['css/style.scss', 'css/solarizeddark.scss', 'css/nanoscroller.css'],
                tasks: ['sass', 'cssmin', 'rev']
            }
        },
        uglify: {
            target: {
                files: {
                    'js/www.js': ['js/d3.js', 'js/jquery-2.0.3.min.js', 'jquery.nanoscroller.min.js', 'js/scripts.js', 'js/topojson.v1.min.js']
                }
            }
        },
        rev: {
            options: {
                algorithm: 'md5',
                length: 16
            },
            assets: {
                files: [{
                    src: ['js/www.js', 'css/www.css']
                }]
            }
        },
        sass: {
            target: {
                files: {
                    'css/style.css': 'css/style.scss',
                    'css/solarizeddark.css': 'css/solarizeddark.scss'
                }
            }
        },
        cssmin: {
            combine: {
                files: {
                    'css/www.css': ['css/style.css', 'css/solarizeddark.css', 'css/nanoscroller.css']
                }
            }
        }
    });

    grunt.loadNpmTasks('grunt-contrib-jshint');
    grunt.loadNpmTasks('grunt-contrib-watch');
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-rev');
    grunt.loadNpmTasks('grunt-sass');
    grunt.loadNpmTasks('grunt-contrib-cssmin');

    grunt.registerTask('default', ['jshint', 'uglify', 'sass', 'cssmin', 'rev']);

};