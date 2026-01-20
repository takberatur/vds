package com.agcforge.videodownloader.utils

class AppManager private constructor() {
    companion object {
        @Volatile
        private var instance: AppManager? = null

        fun getInstance(): AppManager {
            return instance ?: synchronized(this) {
                instance ?: AppManager().also { instance = it }
            }
        }
    }
}